// File: static/js/inout.js

document.addEventListener("DOMContentLoaded", () => {
  const MAX_ROWS       = 10;
  const inoutBtn       = document.getElementById("inoutBtn");
  const inoutForm      = document.getElementById("inoutForm");
  const existingNames  = document.getElementById("existingNames");
  const newNameInput   = document.getElementById("newName");
  const oroshiInput    = document.getElementById("oroshiCode");
  const addClientBtn   = document.getElementById("addClientBtn");
  const clearFormBtn   = document.getElementById("clearFormBtn");
  const submitBtn      = document.getElementById("submitInoutBtn");
  const dateInput      = document.getElementById("inoutDate");
  const slipInput      = document.getElementById("inoutSlipNo");
  const body           = document.getElementById("inoutBody");
  const taxRateInput   = document.getElementById("taxRate");
  const subtotalCell   = document.getElementById("subtotal");
  const totalTaxCell   = document.getElementById("totalTax");
  const grandTotalCell = document.getElementById("grandTotal");

  const modal          = document.getElementById("drugModal");
  const closeModalBtn  = document.getElementById("drugModalClose");
  const searchName     = document.getElementById("searchName");
  const searchSpec     = document.getElementById("searchSpec");
  const searchBtn      = document.getElementById("searchBtn");
  const resultsTbody   = document.querySelector("#searchResults tbody");

  let currentRow = null;
  let taniMap    = {};

  // 外側フォームの submit を止める
  inoutForm.addEventListener("submit", e => e.preventDefault());

  // モーダル内の Enter で親フォーム送信を阻止
  [searchName, searchSpec].forEach(el => {
    el.addEventListener("keydown", e => {
      if (e.key === "Enter") {
        e.preventDefault();
        searchBtn.click();
      }
    });
  });

  // 単位マップ取得
  fetch("/api/tani")
    .then(res => res.json())
    .then(m => taniMap = m)
    .catch(console.error);

  // 得意先一覧ロード
  async function loadClients() {
    existingNames.innerHTML = `<option value="">── 選択 ──</option>`;
    const list = await (await fetch("/api/inout")).json();
    list.forEach(r => {
      const o = document.createElement("option");
      o.value = r.name; o.textContent = r.name;
      existingNames.appendChild(o);
    });
  }

  // 明細行初期化
  function initRows() {
    body.innerHTML = "";
    for (let i = 1; i <= MAX_ROWS; i++) {
      const tr = document.createElement("tr");
      tr.dataset.line = i;
      tr.innerHTML = `
        <td>${i}</td>
        <td class="yj-code"></td>
        <td><input class="jan" type="text"></td>
        <td class="item-name" style="cursor:pointer;">（商品名クリック）</td>
        <td><input class="packaging" readonly></td>
        <td><input class="qty" type="number" min="0"></td>
        <td class="amount">0</td>
        <td class="tax-amount">0</td>
        <td><input class="expiryDate" type="date"></td>
        <td><input class="lotNumber" type="text"></td>
      `;

      // 商品名クリックで検索モーダル
      tr.querySelector(".item-name").addEventListener("click", () => {
        currentRow = tr;
        resultsTbody.innerHTML = "";
        searchName.value = "";
        searchSpec.value = "";
        modal.classList.remove("hidden");
      });

      // 数量・期限入力で再計算
      tr.querySelector(".qty").addEventListener("input", recalcAll);
      tr.querySelector(".expiryDate").addEventListener("change", recalcAll);

      body.appendChild(tr);
    }
  }

  // 再計算ロジック
  function recalcAll() {
    const rate = parseFloat(taxRateInput.value) || 0;
    let sumNet = 0, sumTax = 0;

    body.querySelectorAll("tr").forEach(row => {
      const qtyIn   = parseFloat(row.querySelector(".qty").value) || 0;
      const janQty  = parseFloat(row.dataset.num)   || 0;
      const baseY   = parseFloat(row.dataset.baseY) || 0;
      const realQty = janQty * qtyIn;
      const net     = Math.round(baseY * realQty);
      const tax     = Math.round(net * rate / 100);

      row.querySelector(".amount").textContent     = net;
      row.querySelector(".tax-amount").textContent = tax;
      sumNet += net;
      sumTax += tax;
    });

    subtotalCell.textContent = sumNet;
    totalTaxCell.textContent = sumTax;
    grandTotalCell.textContent = sumNet + sumTax;
  }



  // 得意先登録
  addClientBtn.addEventListener("click", async () => {
    const name = newNameInput.value.trim() || existingNames.value;
    if (!name) {
      alert("得意先を入力または選択してください");
      return;
    }
    await fetch("/api/inout", {
      method: "POST",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify({ name, oroshicode: oroshiInput.value.trim() })
    });
    await loadClients();
  });

  // フォームクリア
  clearFormBtn.addEventListener("click", () => {
    [dateInput, slipInput, existingNames, newNameInput, oroshiInput]
      .forEach(el => el.value = "");
    body.querySelectorAll("input").forEach(i => i.value = "");
    recalcAll();
  });

  // モーダル閉じる
  closeModalBtn.addEventListener("click", () => modal.classList.add("hidden"));

  // 薬品検索
  searchBtn.addEventListener("click", async () => {
    const name = encodeURIComponent(searchName.value.trim());
    const spec = encodeURIComponent(searchSpec.value.trim());
    const list = await (await fetch(`/api/inout/search?name=${name}&spec=${spec}`)).json();
    resultsTbody.innerHTML = "";

    list.forEach(item => {
      const tr = document.createElement("tr");
      const baseY = item.unitYaku / (item.packTotal / item.coef);
      tr.dataset.baseY = baseY.toFixed(6);
      tr.dataset.num   = item.packQtyNumber;
      tr.dataset.code  = item.packQtyUnitCode;
      tr.dataset.unit  = item.unitName;

      const mapped = taniMap[item.packQtyUnitCode] || item.unitName;
      const suffix = item.packQtyUnitCode === 0 ? "" : `/${mapped}`;
      const pkgStr = `${item.packQtyNumber}${item.unitName}${suffix}`;

      tr.innerHTML = `
        <td>${item.yj}</td>
        <td>${item.jan}</td>
        <td>${item.name}</td>
        <td>${item.spec}</td>
        <td>${pkgStr}</td>
        <td>${baseY.toFixed(3)}</td>
      `;

      tr.addEventListener("click", () => {
        currentRow.querySelector(".yj-code").textContent   = item.yj;
        currentRow.querySelector(".jan").value             = item.jan;
        currentRow.querySelector(".item-name").textContent = item.name;
        currentRow.querySelector(".packaging").value       = pkgStr;
        currentRow.dataset.baseY = tr.dataset.baseY;
        currentRow.dataset.num   = tr.dataset.num;
        currentRow.dataset.code  = tr.dataset.code;
        currentRow.dataset.unit  = tr.dataset.unit;
        modal.classList.add("hidden");
        recalcAll();
      });

      resultsTbody.appendChild(tr);
    });
  });

  // 明細送信
  submitBtn.addEventListener("click", async () => {
    if (!dateInput.value || !slipInput.value) {
      alert("日付と伝票番号を入力してください");
      return;
    }
    const type = document.querySelector('input[name="inoutType"]:checked').value;
    const payload = [];

    const rawDate = dateInput.value;               // "2025-12-25"
    const iodDate = rawDate.replace(/-/g, "");     // "20251225"

    body.querySelectorAll("tr").forEach(row => {
      const jan    = row.querySelector(".jan").value;
      const qtyIn  = parseFloat(row.querySelector(".qty").value) || 0;
      const janQty = parseFloat(row.dataset.num) || 0;
      if (!jan || qtyIn === 0) return;

      const realQty = janQty * qtyIn;
      const baseY   = parseFloat(row.dataset.baseY) || 0;
      const net     = Math.round(baseY * realQty);
      const janUnit = taniMap[row.dataset.code] || row.dataset.unit;

      payload.push({
        iodJan:           jan,
        iodDate:          iodDate,
        iodType:          type,
        iodJanQuantity:   janQty,
        iodJanUnit:       janUnit,
        iodQuantity:      realQty,
        iodUnit:          row.dataset.unit,
        iodPackaging:     row.querySelector(".packaging").value,
        iodUnitPrice:     baseY,
        iodSubtotal:      net,
        iodExpiryDate:    row.querySelector(".expiryDate").value.replace(/-/g, ""),
        iodLotNumber:     row.querySelector(".lotNumber").value,
        iodOroshiCode:    oroshiInput.value.trim(),
        iodReceiptNumber: slipInput.value.replace(/-/g, ""),
        iodLineNumber:    parseInt(row.dataset.line, 10),
      });
    });

    console.log("DEBUG payload:", payload);
    await fetch("/api/inout/save", {
      method: "POST",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify(payload)
    });
    alert("保存しました");
  });

  // フォーム トグル
  inoutBtn.addEventListener("click", async () => {
    inoutForm.classList.toggle("hidden");
    if (!inoutForm.classList.contains("hidden")) {
      await loadClients();
      initRows();
      recalcAll();
    }
  });

  // 初期化
  initRows();
  recalcAll();
});