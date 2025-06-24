document.addEventListener('DOMContentLoaded', () => {
  // 必要なDOM要素の取得
  const btn       = document.getElementById('datBtn');
  const input     = document.getElementById('datInput');
  const indicator = document.getElementById('indicator');
  const table     = document.getElementById('outputTable');
  const thead     = table.querySelector('thead');
  const tbody     = table.querySelector('tbody');

  // 「DATファイルアップロード」ボタン押下時：
  // ①テーブル内（出力）の既存内容をクリア
  // ②隠しファイル入力をトリガーしてファイル選択ダイアログを表示
  btn.addEventListener('click', () => {
    thead.innerHTML = "";
    tbody.innerHTML = "";
    input.click();
  });

  // ファイルが選択された時の処理
  input.addEventListener('change', async () => {
    if (!input.files.length) return;

    // アップロード開始状態の表示
    indicator.textContent = 'DATアップロード中…';

    // テーブルヘッダー、ボディの初期化（DAT用）
    thead.innerHTML = `
      <tr>
        <th>卸コード</th>
        <th>日付</th>
        <th>納品／返品</th>
        <th>伝票番号</th>
        <th>行番号</th>
        <th>JANコード</th>
        <th>商品名</th>
        <th>数量</th>
        <th>単価</th>
        <th>小計</th>
        <th>包装薬価</th>
        <th>有効期限</th>
        <th>ロット番号</th>
      </tr>`;
    tbody.innerHTML = '';

    // 選択された各ファイルに対してアップロード処理を実施
    for (let file of input.files) {
      const formData = new FormData();
      formData.append('datFileInput[]', file);

      try {
        // サーバの /uploadDat エンドポイントに対してPOSTリクエスト
        const res = await fetch('/uploadDat', {
          method: 'POST',
          body: formData
        });
        if (!res.ok) {
          throw new Error(`HTTPステータス: ${res.status}`);
        }
        // レスポンスはJSON形式で取得（DAT読み込み件数などの情報を含む）
        const result = await res.json();

        indicator.textContent = `${file.name}: DAT読み込み: ${result.DATReadCount} 件 | MA0作成: ${result.MA0CreatedCount} 件 | 重複: ${result.DuplicateCount} 件`;

        // 結果のDATRecordsをテーブルに追加
        if (result.DATRecords && result.DATRecords.length > 0) {
          result.DATRecords.forEach(record => {
            const tr = document.createElement('tr');
            tr.innerHTML = `
              <td>${record.DatOroshiCode}</td>
              <td>${record.DatDate}</td>
              <td>${record.DatDeliveryFlag}</td>
              <td>${record.DatReceiptNumber}</td>
              <td>${record.DatLineNumber}</td>
              <td>${record.DatJanCode}</td>
              <td>${record.DatProductName}</td>
              <td>${record.DatQuantity}</td>
              <td>${record.DatUnitPrice}</td>
              <td>${record.DatSubtotal}</td>
              <td>${record.DatPackagingDrugPrice}</td>
              <td>${record.DatExpiryDate}</td>
              <td>${record.DatLotNumber}</td>
            `;
            tbody.appendChild(tr);
          });
        }
      } catch (err) {
        indicator.textContent = "アップロード中にエラーが発生しました: " + err.message;
        console.error(err);
      }
    }
    indicator.textContent += " 完了";
    // 次回のアップロードのためにファイル入力をリセット
    input.value = '';
  });
});