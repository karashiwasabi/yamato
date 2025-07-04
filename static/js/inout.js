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
   // --- 変更後 ---
  addClientBtn.addEventListener("click", async () => {
    // 1) 得意先名取得
    const name = newNameInput.value.trim() || existingNames.value;
    if (!name) {
      return alert("得意先を入力または選択してください");
    }

    // 2) サーバーへ登録リクエスト
    const res = await fetch("/api/inout", {
      method:  "POST",
      headers: { "Content-Type": "application/json" },
      body:    JSON.stringify({
        name,
        oroshicode: oroshiInput.value.trim()
      })
    });

    // 3) 成否アラート
    if (!res.ok) {
      return alert("得意先登録に失敗しました");
    }
    alert("得意先を登録しました");

    // 4) フィールドクリア＋一覧再ロード
    newNameInput.value = "";
    oroshiInput.value  = "";
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

submitBtn.addEventListener("click", async () => {
  // 1) 日付＆伝票番号の必須チェック
  if (!dateInput.value || !slipInput.value) {
    alert("日付と伝票番号を入力してください");
    return;
  }

  // 2) 出庫／入庫タイプを数値コードに変換（出庫→3、入庫→4）
  const rawType  = document.querySelector('input[name="inoutType"]:checked').value;
  const typeCode = rawType === "出庫" ? 3 : 4;

  // 3) 伝票番号からハイフンを除去
  const slipNo = slipInput.value.trim().replace(/-/g, "");

  // 4) 日付を YYYYMMDD 形式に変換
  const iodDate = dateInput.value.replace(/-/g, "");

  // 5) 明細ごとに DTO を組み立て
  const payload = [];
  body.querySelectorAll("tr").forEach(row => {
    const jan     = row.querySelector(".jan").value.trim();
    
    const qtyIn   = parseFloat(row.querySelector(".qty").value) || 0;
    if (!jan || qtyIn === 0) return;  // JAN が空 or 数量０ はスキップ

    // data 属性から取り出す
    const janQty     = parseFloat(row.dataset.num)   || 0;   // １パックあたりのJAN数量
    const baseY      = parseFloat(row.dataset.baseY) || 0;   // 薬価（税抜）
    const unitCode   = row.dataset.code    || "";            // 包装単位コード or 名称
    const unitName   = row.dataset.unit    || "";            // 包装単位名称
    const packaging  = row.querySelector(".packaging").value.trim();
    const rawExp     = row.querySelector(".expiryDate").value;
    const expDate    = rawExp ? rawExp.replace(/-/g, "") : "";
    const lotNo      = row.querySelector(".lotNumber").value.trim();

    // 実 JAN 数量／金額計算
    const realQty = janQty * qtyIn;
    const netAmt  = Math.round(baseY * realQty);   // 税抜金額

    payload.push({
      iodJan:           jan,          // JANコード
      iodProductName:   row.querySelector(".item-name").textContent,  // 追加
      iodDate:          iodDate,      // 登録日
      iodType:          typeCode,     // 出庫(3)/入庫(4)
      iodJanQuantity:   qtyIn,        // 入力パック数
      iodJanUnit:       unitCode,     // 包装単位コード or 名称
      iodQuantity:      realQty,      // 実JAN数量
      iodUnit:          unitName,     // 包装単位名称
      iodPackaging:     packaging,    // パッケージ文字列
      iodUnitPrice:     baseY,        // １JANあたり薬価(税抜)
      iodSubtotal:      netAmt,       // 小計(税抜)
      iodExpiryDate:    expDate,      // 有効期限(YYYYMMDD)
      iodLotNumber:     lotNo,        // ロット番号
      iodOroshiCode:    oroshiInput.value.trim(),  // 卸コード
      iodReceiptNumber: slipNo,       // 伝票番号
      iodLineNumber:    parseInt(row.dataset.line, 10) // 行番号
    });
  });

  // 6) サーバーへ送信
  try {
    const res = await fetch("/api/inout/save", {
      method:  "POST",
      headers: { "Content-Type": "application/json" },
      body:    JSON.stringify(payload)
    });
    if (!res.ok) throw new Error(res.statusText);

    alert("保存しました");
    location.reload();

  } catch (err) {
    console.error(err);
    alert("登録エラー: " + err.message);
  }
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