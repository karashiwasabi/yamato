document.addEventListener('DOMContentLoaded', () => {
  const ma0Btn = document.getElementById('ma0Btn');
  const ma0Container = document.getElementById('ma0Container');
  const indicator = document.getElementById('indicator');

  ma0Btn.addEventListener('click', async () => {
    indicator.textContent = 'MA0表示中…';
    try {
      const res = await fetch('/viewMA0');
      if (!res.ok) throw new Error(`HTTPステータス: ${res.status}`);

      // サーバから返却される JSON データは、"MA0Records" キーに MA0Record の配列が入っているものとする
      const data = await res.json();

      if (data.MA0Records && data.MA0Records.length > 0) {
        ma0Container.innerHTML =
          `<h2>MA0レコード（${data.MA0Records.length} 件）</h2>` +
          `<ul>` +
          data.MA0Records
            .map(rec => `<li>${rec.janCode} (包装単位: ${rec.packagingUnit}, 換算係数: ${rec.conversionFactor})</li>`)
            .join('') +
          `</ul>`;
      } else {
        ma0Container.innerHTML = `<h2>MA0は空です。</h2>`;
      }
      indicator.textContent = '';
    } catch (err) {
      indicator.textContent = "MA0の表示中にエラー: " + err.message;
      console.error(err);
    }
  });
});