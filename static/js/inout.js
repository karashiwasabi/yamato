// File: static/js/inout.js

document.addEventListener("DOMContentLoaded", () => {
  // Form & modal elements
  const inoutBtn      = document.getElementById("inoutBtn");
  const inoutForm     = document.getElementById("inoutForm");
  const existingNames = document.getElementById("existingNames");
  const newNameInput  = document.getElementById("newName");
  const oroshiInput   = document.getElementById("oroshiCode");
  const addClientBtn  = document.getElementById("addClientBtn");

  const inoutTable    = document.getElementById("inoutTable");
  const modal         = document.getElementById("drugModal");
  const closeBtn      = document.getElementById("drugModalClose");
  const searchName    = document.getElementById("searchName");
  const searchSpec    = document.getElementById("searchSpec");
  const searchBtn     = document.getElementById("searchBtn");
  const resultsTbody  = document.querySelector("#searchResults tbody");

  // Summary & tax cells
  const subtotalCell = document.getElementById("subtotal");
  const totalTaxCell = document.getElementById("totalTax");
  const grandTotalCell = document.getElementById("grandTotal");
  const taxRateInput = document.getElementById("taxRate");

  let currentRow = null;
  const taniMap = {};

  // Load unit reverse map
  fetch("/api/tani")
    .then(r => r.json())
    .then(m => Object.assign(taniMap, m))
    .catch(console.error);

  // Load existing clients
  async function loadNames() {
    existingNames.innerHTML = `<option value="">──選択──</option>`;
    try {
      const res = await fetch("/api/inout");
      const list = await res.json();
      list.forEach(r => {
        const o = document.createElement("option");
        o.value = r.name; o.textContent = r.name;
        existingNames.appendChild(o);
      });
    } catch(e) { console.error(e); }
  }

  // Register client
  addClientBtn.addEventListener("click", async () => {
    const name = newNameInput.value.trim() || existingNames.value;
    if(!name) return alert("得意先を入力または選択してください");
    try {
      const res = await fetch("/api/inout", {
        method:"POST",
        headers:{"Content-Type":"application/json"},
        body:JSON.stringify({name,oroshicode:oroshiInput.value.trim()})
      });
      if(!res.ok) throw new Error(await res.text());
      alert("登録完了");
      newNameInput.value=""; oroshiInput.value="";
      existingNames.value="";
      await loadNames();
    } catch(e) {
      console.error(e); alert("登録失敗："+e.message);
    }
  });

  // Toggle form
  inoutBtn.addEventListener("click", () => {
    inoutForm.classList.toggle("hidden");
    if(!inoutForm.classList.contains("hidden")) {
      loadNames(); recalcAll();
    }
  });

  // Show modal on product-name click
  inoutTable.addEventListener("click", e => {
    const td = e.target.closest("td.item-name");
    if(!td) return;
    currentRow = td.closest("tr");
    resultsTbody.innerHTML="";
    searchName.value=""; searchSpec.value="";
    modal.classList.remove("hidden");
    searchName.focus();
  });

  closeBtn.addEventListener("click", () => modal.classList.add("hidden"));

  // Search & render modal rows
  searchBtn.addEventListener("click", async () => {
    const n = encodeURIComponent(searchName.value.trim());
    const s = encodeURIComponent(searchSpec.value.trim());
    try {
      const res = await fetch(`/api/inout/search?name=${n}&spec=${s}`);
      if(!res.ok) throw new Error(res.statusText);
      const list = await res.json();
      resultsTbody.innerHTML = "";

      list.forEach(item => {
        const baseY = item.unitYaku/(item.packTotal/item.coef);
        const code  = item.packQtyUnitCode;
        const num   = item.packQtyNumber;
        const jc39  = item.unitName;
        const mapped = taniMap[code];
        const suffix = code===0?"":`/${mapped||jc39}`;
        const pkgText = `${num}${jc39}${suffix}`;

        const tr = document.createElement("tr");
        tr.innerHTML = `
          <td>${item.yj}</td>
          <td>${item.jan}</td>
          <td>${item.name}</td>
          <td>${item.spec}</td>
          <td>${pkgText}</td>
          <td>${baseY.toFixed(3)}</td>
        `;
        tr.dataset.baseY = baseY;
        tr.dataset.num   = num;
        tr.dataset.code  = code;
        tr.dataset.unit  = jc39;

        tr.addEventListener("click", () => {
          currentRow.querySelector(".yj-code").textContent   = item.yj;
          currentRow.querySelector(".jan").value             = item.jan;
          currentRow.querySelector(".item-name").textContent = item.name;
          currentRow.querySelector(".packaging").value       = pkgText;
          currentRow.dataset.baseY = baseY;
          currentRow.dataset.num   = num;
          currentRow.dataset.code  = code;
          currentRow.dataset.unit  = jc39;
          modal.classList.add("hidden");
          recalcAll();
        });

        resultsTbody.appendChild(tr);
      });
    } catch(e) {
      console.error(e); alert("検索失敗："+e.message);
    }
  });

  // Recalculate line-item rounding then tax
  function recalcAll() {
    const rate = parseFloat(taxRateInput.value)||0;
    let sumAmt=0, sumTax=0;

    inoutTable.querySelectorAll("tbody tr").forEach(row=>{
      if(["subtotalRow","taxRow"].includes(row.id)) return;
      const qty  = parseInt(row.querySelector(".qty").value,10)||0;
      const base = parseFloat(row.dataset.baseY)||0;
      const num  = parseFloat(row.dataset.num)||0;
      const code = +row.dataset.code;

      // raw net
      const rawNet = code===0
        ? base*qty
        : base*num*qty;
      // rounded net
      const net = Math.round(rawNet);
      row.querySelector(".amount").textContent = net;

      // line tax
      const tax = Math.round(net*rate/100);
      row.querySelector(".tax-amount").textContent = tax;

      sumAmt += net;
      sumTax += tax;
    });

    // subtotal & total
    subtotalCell.textContent = sumAmt;
    totalTaxCell.textContent = sumTax;
    grandTotalCell.textContent = sumAmt + sumTax;
  }

  // Listen inputs
  inoutTable.addEventListener("input", e => {
    if(e.target.classList.contains("qty")) recalcAll();
  });
  taxRateInput.addEventListener("input", recalcAll);

  // initial
  recalcAll();
});