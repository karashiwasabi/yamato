// File: static/js/aggregate.js
document.addEventListener("DOMContentLoaded", () => {
  const btn    = document.getElementById("aggregateBtn");
  const filterDiv = document.getElementById("aggregateFilter");
  const form   = document.getElementById("filterForm");
  const indicator = document.getElementById("indicator");
  const table     = document.getElementById("outputTable");
  const thead     = table.querySelector("thead");
  const tbody     = table.querySelector("tbody");

  // 初期化
  thead.innerHTML = "";
  tbody.innerHTML = "";

  btn.addEventListener("click", () => {
    filterDiv.style.display = "block";
    indicator.textContent   = "";
    thead.innerHTML = "";
    tbody.innerHTML = "";
  });

  form.addEventListener("submit", async e => {
    e.preventDefault();
    thead.innerHTML = "";
    tbody.innerHTML = "";

    const from   = form.elements["from"].value;
    const to     = form.elements["to"].value;
    const filter = form.elements["filter"].value.trim();
    if (!from || !to) {
      alert("開始日と終了日を指定してください");
      return;
    }

    const params = new URLSearchParams({ from, to });
    if (filter) params.append("filter", filter);
    ["doyaku","gekiyaku","mayaku"].forEach(name => {
      if (form.elements[name].checked) params.append(name, "1");
    });
    ["kakuseizai","kakuseizaiGenryou"].forEach(name => {
      if (form.elements[name].checked) params.append(name, "1");
    });
    const ks = ["kousei1","kousei2","kousei3"]
      .filter(id => form.elements[id].checked)
      .map(id => form.elements[id].value);
    if (ks.length) params.append("kouseishinyaku", ks.join(","));

    const url = `/aggregate?${params.toString()}`;
    console.log("▶ Fetching:", url);
    indicator.textContent = `集計中… (${from} ～ ${to})`;

    let groups;
    try {
      const res = await fetch(url);
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

    for (const yj of Object.keys(groups)) {
      let productName = "";
      if (yj) {
        try {
          const pr = await fetch(`/productName?yj=${encodeURIComponent(yj)}`);
          if (pr.ok) {
            const pj = await pr.json();
            productName = pj.productName || "";
          }
        } catch {}
      }
      // YJ 見出し
      const trHead = document.createElement("tr");
      trHead.innerHTML = `
        <td colspan="13">
          YJコード: ${yj}${productName ? " " + productName : ""}
        </td>`;
      tbody.appendChild(trHead);

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
          <td>${d.lineNumber}</td>`;
        tbody.appendChild(tr);
      });
    }

    indicator.textContent = `集計完了 (${from} ～ ${to})`;
  });
});