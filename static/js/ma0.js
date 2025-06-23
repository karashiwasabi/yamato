// File: static/js/ma0.js
document.addEventListener('DOMContentLoaded', () => {
  const ma0Btn       = document.getElementById('ma0Btn');
  const ma0Container = document.getElementById('ma0Container');
  const indicator    = document.getElementById('indicator');

  ma0Btn.addEventListener('click', async () => {
    indicator.textContent = 'MA0表示中…';
    ma0Container.innerHTML = ''; // 前回の結果をクリア

    try {
      const res = await fetch('/viewMA0');
      if (!res.ok) {
        throw new Error(`HTTPステータス: ${res.status}`);
      }

      const data = await res.json();
      // サーバーは JSON 配列を返す想定
      const list = Array.isArray(data) ? data : [];

      if (list.length > 0) {
        ma0Container.innerHTML =
          `<h2>MA0レコード（${list.length} 件）</h2>` +
          `<ul>` +
          list.map(rec =>
            `<li>` +
              `<strong>JANコード:</strong> ${rec.mA000JC000JanCode}<br>` +
              `<strong>YJコード:</strong> ${rec.mA009JC009YJCode || '―'}<br>` +
              `<strong>包装単位コード:</strong> ${rec.mA131JA006HousouSuuryouSuuchi || '―'}` +
            `</li>`
          ).join('') +
          `</ul>`;
      } else {
        ma0Container.innerHTML = `<h2>MA0は空です。</h2>`;
      }
    } catch (err) {
      ma0Container.innerHTML = `<h2>MA0の表示中にエラーが発生しました。</h2>`;
      console.error('MA0 fetch error:', err);
    } finally {
      indicator.textContent = '';
    }
  });
});