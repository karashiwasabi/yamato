// File: static/js/aggregate.js
document.addEventListener("DOMContentLoaded", () => {
  const btnAgg     = document.getElementById("aggregateBtn");
  const filterDiv  = document.getElementById("aggregateFilter");
  const formFilter = document.getElementById("filterForm");
  const indicator  = document.getElementById("indicator");
  const table      = document.getElementById("outputTable");
  const tbody      = table.querySelector("tbody");

  // 初期化
  table.querySelector("thead").innerHTML = "";
  tbody.innerHTML                        = "";

  // 「集計」ボタンを押したらフィルタを表示＆クリア
  btnAgg.addEventListener("click", () => {
    filterDiv.style.display = "block";
    indicator.textContent   = "";
    table.querySelector("thead").innerHTML = "";
    tbody.innerHTML                        = "";
  });

  // フィルタ実行
  formFilter.addEventListener("submit", async e => {
    e.preventDefault();
    indicator.textContent = "";

    // テーブル初期化
    table.querySelector("thead").innerHTML = "";
    tbody.innerHTML                        = "";

    // パラメータ取得
    const from   = formFilter.elements["from"].value;
    const to     = formFilter.elements["to"].value;
    const filter = formFilter.elements["filter"].value.trim();
    if (!from || !to) {
      alert("開始日と終了日を指定してください");
      return;
    }

    // URL組み立て
    const params = new URLSearchParams({ from, to });
    if (filter) params.append("filter", filter);
    ["doyaku","gekiyaku","mayaku"].forEach(k => {
      if (formFilter.elements[k].checked)
        params.append(k, formFilter.elements[k].value);
    });
    ["kakuseizai","kakuseizaiGenryou"].forEach(k => {
      if (formFilter.elements[k].checked)
        params.append(k, formFilter.elements[k].value);
    });
    const ks = ["kousei1","kousei2","kousei3"]
      .filter(id => formFilter.elements[id].checked)
      .map(id => formFilter.elements[id].value);
    if (ks.length) params.append("kouseishinyaku", ks.join(","));

    indicator.textContent = `集計中… (${from}～${to})`;

    let groups;
    try {
      const res = await fetch(`/aggregate?${params.toString()}`);
      if (!res.ok) throw new Error(res.statusText);
      groups = await res.json();
    } catch (err) {
      indicator.textContent = `集計失敗: ${err.message}`;
      return;
    }
    if (!groups || !Object.keys(groups).length) {
      indicator.textContent = "該当データがありません";
      return;
    }

    // グループごとに YJ見出し＋列ヘッダー＋明細を描画
    for (const yj of Object.keys(groups)) {
      // 1) 商品名フェッチ
      let productName = "";
      if (yj) {
        try {
          const pr = await fetch(`/productName?yj=${encodeURIComponent(yj)}`);
          if (pr.ok) {
            const pj = await pr.json();
            productName = pj.productName || "";
          }
        } catch {/* ignore */}
      }

      // 2) YJコード見出し行
      const trGroup = document.createElement("tr");
      trGroup.innerHTML = `
        <td colspan="13">
          YJコード: ${yj}${productName ? " " + productName : ""}
        </td>`;
      tbody.appendChild(trGroup);

      // 3) 列ヘッダー行
      const trCols = document.createElement("tr");
      trCols.innerHTML = `
        <th>日付</th><th>種類</th><th>数量</th>
        <th>単位</th><th>包装</th><th>個数</th>
        <th>単価</th><th>金額</th><th>期限</th>
        <th>ロット</th><th>卸コード</th>
        <th>伝票番号</th><th>行番号</th>
      `;
      tbody.appendChild(trCols);

      // 4) 明細行
      groups[yj].forEach(d => {
        const tr = document.createElement("tr");
        tr.innerHTML = `
          <td>${d.date}</td>
          <td>${d.type}</td>
          <td>${d.quantity}</td>
          <td>${d.unit}</td>
          <td>${d.packaging}</td>
          <td>${d.count}</td>
          <td>${d.unitPrice}</td>
          <td>${d.subtotal}</td>
          <td>${d.expiryDate}</td>
          <td>${d.lotNumber}</td>
          <td>${d.oroshiCode}</td>
          <td>${d.receiptNumber}</td>
          <td>${d.lineNumber}</td>
        `;
        tbody.appendChild(tr);
      });
    }

    indicator.textContent = `集計完了 (${from}～${to})`;
  });
});