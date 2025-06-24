// File: static/js/aggregate.js
// Handles “集計” → 期間指定 → /aggregate → テーブル描画 → 商品名フェッチ

document.addEventListener('DOMContentLoaded', () => {
  const aggBtn    = document.getElementById('aggregateBtn');
  const indicator = document.getElementById('indicator');
  const table     = document.getElementById('outputTable');
  const thead     = table.querySelector('thead');
  const tbody     = table.querySelector('tbody');

  aggBtn.addEventListener('click', () => {
    // テーブル＆インジケーターをクリア
    thead.innerHTML = '';
    tbody.innerHTML = '';
    indicator.innerHTML = `
      <input type="date" id="fromDate"> ～
      <input type="date" id="toDate">
      <button id="executeAggregate">実行</button>
    `;

    const execBtn = document.getElementById('executeAggregate');
    execBtn.addEventListener('click', async () => {
      const from = document.getElementById('fromDate').value;
      const to   = document.getElementById('toDate').value;
      if (!from || !to) {
        return alert('期間を指定してください');
      }

      indicator.textContent = `集計中… (${from} ～ ${to})`;

      // 1) 集計データ取得
      let groups;
      try {
        const res = await fetch(`/aggregate?from=${from}&to=${to}`);
        if (!res.ok) throw new Error(`HTTP ${res.status}`);
        groups = await res.json();
      } catch (err) {
        indicator.textContent = `集計失敗: ${err.message}`;
        console.error(err);
        return;
      }

      // 2) テーブルヘッダー定義
      thead.innerHTML = `
        <tr>
          <th>日付</th>
          <th>種類</th>
          <th>数量</th>
          <th>単位</th>
          <th>包装</th>
          <th>個数</th>
          <th>単価</th>
          <th>金額</th>
          <th>期限</th>
          <th>ロット</th>
          <th>卸コード</th>
          <th>伝票番号</th>
          <th>行番号</th>
        </tr>`;

      // 3) YJごとに描画
      for (const yj in groups) {
        // 商品名を取得
        let productName = '';
        try {
          const r = await fetch(`/productName?yj=${encodeURIComponent(yj)}`);
          if (r.ok) {
            const j = await r.json();
            productName = j.productName || '';
          }
        } catch (e) {
          console.warn('productName fetch error', e);
        }

        // グループ見出し行
        const headerRow = document.createElement('tr');
        headerRow.innerHTML = `
          <td colspan="13">
            YJコード: ${yj}
            ${productName ? ` ${productName}` : ''}
          </td>`;
        tbody.appendChild(headerRow);

        // 明細行
        groups[yj].forEach(d => {
          const tr = document.createElement('tr');
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

      indicator.textContent = '集計完了';
    }, { once: true });
  });
});