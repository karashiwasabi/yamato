document.addEventListener("DOMContentLoaded", () => {
  const aggregateBtn = document.getElementById("aggregateBtn");
  const printBtn     = document.getElementById("printTable");
  const filterDiv    = document.getElementById("aggregateFilter");
  const formFilter   = document.getElementById("filterForm");
  const indicator    = document.getElementById("indicator");
  const table        = document.getElementById("outputTable");
  const thead        = table.querySelector("thead");
  const tbody        = table.querySelector("tbody");

  // 初期化
  thead.innerHTML = "";
  tbody.innerHTML = "";

  // 印刷ボタン
  if (printBtn) {
    printBtn.addEventListener("click", () => window.print());
  }

  // 集計ボタン
  aggregateBtn.addEventListener("click", () => {
    filterDiv.style.display = "block";
    indicator.textContent   = "";
    thead.innerHTML         = "";
    tbody.innerHTML         = "";
  });

  // フィルタ実行
  formFilter.addEventListener("submit", async e => {
    e.preventDefault();
    thead.innerHTML = "";
    tbody.innerHTML = "";

    const from = formFilter.querySelector('input[name="from"]').value;
    const to   = formFilter.querySelector('input[name="to"]').value;
    if (!from || !to) {
      alert("開始日と終了日を指定してください");
      return;
    }
    const filter = formFilter.querySelector('input[name="filter"]').value.trim();

    // クエリ生成
    const params = new URLSearchParams({ from, to });
    if (filter) params.append("filter", filter);
    ["doyaku","gekiyaku","mayaku","kakuseizai","kakuseizaiGenryou"]
      .forEach(name => {
        const cb = formFilter.querySelector(`input[name="${name}"]`);
        if (cb && cb.checked) params.append(name, cb.value);
      });
    const kousei = Array.from(
      formFilter.querySelectorAll('input[name="kouseishinyaku"]:checked')
    ).map(cb => cb.value);
    if (kousei.length) {
      params.append("kouseishinyaku", kousei.join(","));
    }

    indicator.textContent = `集計中… (${from} ～ ${to})`;

    let data;
    try {
      const res = await fetch(`/aggregate?${params.toString()}`);
      if (!res.ok) throw new Error(res.statusText);
      data = await res.json();
    } catch (err) {
      indicator.textContent = `集計失敗: ${err.message}`;
      return;
    }
    if (!data || !Object.keys(data).length) {
      indicator.textContent = "該当データがありません";
      return;
    }

    // 描画: YJ → 包装分類キー → 明細
    Object.entries(data).forEach(([yj, {productName, groups}]) => {
      // YJヘッダ
      const trYJ = document.createElement("tr");
      trYJ.innerHTML = `<td colspan="14">
        YJコード: ${yj}${productName ? " " + productName : ""}
      </td>`;
      tbody.appendChild(trYJ);

      // 各包装分類キーごとに
      Object.entries(groups).forEach(([pk, list]) => {
        // 包装分類ヘッダ
        const trPK = document.createElement("tr");
        trPK.innerHTML = `<td colspan="14">包装分類: ${pk}</td>`;
        tbody.appendChild(trPK);

        // 列ヘッダー
        const trCols = document.createElement("tr");
        trCols.innerHTML = `
          <th>日付</th><th>種類</th><th>数量</th>
          <th>単位</th><th>包装</th><th>個数</th>
          <th>単価</th><th>金額</th><th>期限</th>
          <th>ロット</th><th>卸コード</th>
          <th>伝票番号</th><th>行番号</th>`;
        tbody.appendChild(trCols);

        // 明細
        list.forEach(d => {
          const tr = document.createElement("tr");
          tr.innerHTML = `
            <td>${d.date}</td><td>${d.type}</td>
            <td>${d.quantity}</td><td>${d.unit}</td><td>${d.packaging}</td>
            <td>${d.count}</td><td>${d.unitPrice}</td><td>${d.subtotal}</td>
            <td>${d.expiryDate}</td><td>${d.lotNumber}</td>
            <td>${d.oroshiCode}</td><td>${d.receiptNumber}</td><td>${d.lineNumber}</td>`;
          tbody.appendChild(tr);
        });
      });
    });

    indicator.textContent = `集計完了 (${from} ～ ${to})`;
  });
});