document.addEventListener('DOMContentLoaded', () => {
  const ma0Btn       = document.getElementById('ma0Btn');
  const ma0Container = document.getElementById('ma0Container');
  const indicator    = document.getElementById('indicator');

  ma0Btn.addEventListener('click', async () => {
    indicator.textContent = 'MA0表示中…';
    try {
      const res = await fetch('/ma0/view');
      if (!res.ok) throw new Error(`HTTPステータス: ${res.status}`);

      const data = await res.json();
      const list = Array.isArray(data) ? data : data.MA0Records || [];

      if (list.length > 0) {
        ma0Container.innerHTML =
          `<h2>MA0レコード（${list.length} 件）</h2>` +
          `<ul>` +
          list.map(rec =>
            `<li>${rec.mA000JC000JanCode} (YJコード: ${rec.mA009JC009YJCode}, 単位コード: ${rec.mA131JA007HousouSuuryouTaniCode})</li>`
          ).join('') +
          `</ul>`;
      } else {
        ma0Container.innerHTML = `<h2>MA0は空です。</h2>`;
      }
    } catch (err) {
      // エラー時の表示
      indicator.textContent = 'エラーが発生しました: ' + err.message;
      console.error(err);
    }
    // 必要なら finally で完了後の後片付けも可能
    // finally {
    //   indicator.textContent = '';
    // }
  });
});