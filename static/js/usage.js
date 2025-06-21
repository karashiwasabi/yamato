document.addEventListener('DOMContentLoaded', () => {
  const usageBtn   = document.getElementById('usageBtn');
  const usageInput = document.getElementById('usageInput');
  const indicator  = document.getElementById('indicator');
  const table      = document.getElementById('outputTable');
  const thead      = table.querySelector('thead');
  const tbody      = table.querySelector('tbody');

  usageBtn.addEventListener('click', () => usageInput.click());

  usageInput.addEventListener('change', async () => {
    if (!usageInput.files || usageInput.files.length === 0) return;

    indicator.textContent = 'USAGEアップロード中…';
    // 新しいヘッダー（項目名）に合わせて設定
    thead.innerHTML = `
      <tr>
        <th>日付</th>
        <th>YJコード</th>
        <th>JANコード</th>
        <th>商品名</th>
        <th>数量</th>
        <th>単位コード</th>
        <th>単位名称</th>
      </tr>`;
    tbody.innerHTML = '';

    for (let file of usageInput.files) {
      const formData = new FormData();
      formData.append('usageFileInput[]', file);

      try {
        const res = await fetch('/uploadUsage', { method: 'POST', body: formData });
        if (!res.ok) throw new Error(`HTTPエラー: ${res.status}`);
        const result = await res.json();
        indicator.textContent = `${file.name}: USAGE読み込み: ${result.TotalRecords} 件`;
        if (result.USAGERecords && result.USAGERecords.length > 0) {
          result.USAGERecords.forEach(record => {
            const tr = document.createElement('tr');
            tr.innerHTML = `
              <td>${record.usageDate}</td>
              <td>${record.usageYjCode}</td>
              <td>${record.usageJanCode}</td>
              <td>${record.usageProductName}</td>
              <td>${record.usageAmount}</td>
              <td>${record.usageUnit}</td>
              <td>${record.usageUnitName}</td>
            `;
            tbody.appendChild(tr);
          });
        }
      } catch (error) {
        indicator.textContent = "アップロード中にエラー: " + error.message;
        console.error("USAGEアップロードエラー:", error);
      }
    }
    indicator.textContent += " 完了";
    usageInput.value = '';
  });
});