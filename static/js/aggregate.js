// static/js/aggregate.js

document.addEventListener("DOMContentLoaded", () => {
  const btnAggregate = document.getElementById("aggregateBtn");
  const filterDiv    = document.getElementById("aggregateFilter");
  const formFilter   = document.getElementById("filterForm");
  const indicator    = document.getElementById("indicator");
  const table        = document.getElementById("outputTable");
  const thead        = table.querySelector("thead");
  const tbody        = table.querySelector("tbody");

  // 「集計」ボタンでフィルタ部を表示
  btnAggregate.addEventListener("click", () => {
    filterDiv.style.display = "block";
    indicator.textContent   = "";
    thead.innerHTML         = "";
    tbody.innerHTML         = "";
  });

  // フォーム submit で集計実行
  formFilter.addEventListener("submit", async (e) => {
    e.preventDefault();

    // 日付・テキストフィルタ取得
    const from   = formFilter.elements["from"].value;
    const to     = formFilter.elements["to"].value;
    const filter = formFilter.elements["filter"].value.trim();
    if (!from || !to) {
      alert("開始日と終了日を指定してください");
      return;
    }

    // クエリパラメータ組み立て
    const params = new URLSearchParams({ from, to });
    if (filter) params.append("filter", filter);

    // 毒薬・劇薬・麻薬
    ["doyaku", "gekiyaku", "mayaku"].forEach(key => {
      if (formFilter.elements[key].checked) {
        params.append(key, formFilter.elements[key].value);
      }
    });

    // 覚せい剤・原料
    ["kakuseizai", "kakuseizaiGenryou"].forEach(key => {
      if (formFilter.elements[key].checked) {
        params.append(key, formFilter.elements[key].value);
      }
    });

    // 向精神薬 (複数選択可)
    const kouseiVals = ["kousei1", "kousei2", "kousei3"]
      .filter(id => formFilter.elements[id].checked)
      .map(id => formFilter.elements[id].value);
    if (kouseiVals.length) {
      params.append("kouseishinyaku", kouseiVals.join(","));
    }

    const url = `/aggregate?${params.toString()}`;
    indicator.textContent = `集計中… (${from} ～ ${to})`;
    thead.innerHTML = "";
    tbody.innerHTML = "";

    try {
      const res  = await fetch(url);
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data   = await res.json();
      const groups = data.groups || data;

      if (!groups || Object.keys(groups).length === 0) {
        indicator.textContent = "該当データがありません";
        return;
      }

      // テーブルヘッダー（商品名列を除く13列）
      thead.innerHTML = `
        <tr>
          <th>日付</th><th>種類</th><th>数量</th>
          <th>単位</th><th>包装</th><th>個数</th>
          <th>単価</th><th>金額</th><th>期限</th>
          <th>ロット</th><th>卸コード</th>
          <th>伝票番号</th><th>行番号</th>
        </tr>`;

      // グループ描画：for…of で await を使えるように
      for (const yj of Object.keys(groups)) {
        // 商品名取得
        let productName = "";
        try {
          const pr = await fetch(
            `/productName?yj=${encodeURIComponent(yj)}`
          );
          if (pr.ok) {
            const pj = await pr.json();
            productName = pj.productName || "";
          }
        } catch {
          // ignore
        }

        // グループ見出し (colspan=13)
        const hdr = document.createElement("tr");
        hdr.innerHTML = `
          <td colspan="13">
            YJコード: ${yj}${productName ? " " + productName : ""}
          </td>`;
        tbody.appendChild(hdr);

        // 明細行 (商品名列は出さない)
        for (const d of groups[yj]) {
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
        }
      }

      indicator.textContent = `集計完了 (${from} ～ ${to})`;
    } catch (err) {
      console.error(err);
      indicator.textContent = `集計失敗: ${err.message}`;
    }
  });
});