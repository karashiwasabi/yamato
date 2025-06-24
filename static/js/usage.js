document.addEventListener('DOMContentLoaded', () => {
  // 必要なDOM要素の取得
  const usageBtn   = document.getElementById('usageBtn');
  const usageInput = document.getElementById('usageInput');
  const indicator  = document.getElementById('indicator');
  const table      = document.getElementById('outputTable');
  const thead      = table.querySelector('thead');
  const tbody      = table.querySelector('tbody');

  // 「USAGEファイルアップロード」ボタン押下時：
  // ①出力テーブルの内容をクリア
  // ②隠しファイル入力をトリガー
  usageBtn.addEventListener('click', () => {
    thead.innerHTML = "";
    tbody.innerHTML = "";
    usageInput.click();
  });

  // ファイルが選択された時の処理
  usageInput.addEventListener('change', async () => {
    if (!usageInput.files || usageInput.files.length === 0) return;

    indicator.textContent = 'USAGEアップロード中…';

    // USAGE用のテーブルヘッダー、ボディを初期化
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

    // 選択された各ファイルに対してアップロード処理を実施
    for (let file of usageInput.files) {
      const formData = new FormData();
      formData.append('usageFileInput[]', file);

      try {
        const res = await fetch('/uploadUsage', {
          method: 'POST',
          body: formData
        });
        if (!res.ok) {
          throw new Error(`HTTPエラー: ${res.status}`);
        }
        // サーバからJSON形式で結果を取得
        const result = await res.json();
        indicator.textContent = `${file.name}: USAGE読み込み: ${result.TotalRecords} 件`;

        // 結果のUSAGERecordsをテーブルに追加
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