document.addEventListener('DOMContentLoaded', () => {
  // 各DOM要素の取得
  const btn       = document.getElementById('datBtn');
  const input     = document.getElementById('datInput');
  const indicator = document.getElementById('indicator');
  const table     = document.getElementById('outputTable');
  const thead     = table.querySelector('thead');
  const tbody     = table.querySelector('tbody');

  // DATファイルアップロードボタンをクリックで隠しファイル入力をトリガー
  btn.addEventListener('click', () => {
    input.click();
  });

  // ファイル選択後の処理
  input.addEventListener('change', async () => {
    if (!input.files.length) return;

    // アップロード開始状態を indicator に表示
    indicator.textContent = 'DATアップロード中…';

    // テーブルヘッダー、ボディの初期化
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

    // 選択された全ファイルについてループ処理
    for (let file of input.files) {
      const formData = new FormData();
      formData.append('datFileInput[]', file);

      try {
        // /uploadDat エンドポイントに対して POST リクエスト送信
        const res = await fetch('/uploadDat', { method: 'POST', body: formData });
        if (!res.ok) {
          throw new Error(`HTTPステータス: ${res.status}`);
        }
        // サーバーから返却される JSON を取得（"DATRecords" などのキーを持つ）
        const result = await res.json();

        // 指定ファイルの処理結果を indicator に表示
        indicator.textContent = `${file.name}: DAT読み込み: ${result.DATReadCount} 件 | MA0作成: ${result.MA0CreatedCount} 件 | 重複: ${result.DuplicateCount} 件`;

        // DATRecords に含まれる各レコードをテーブルへ追加
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
        // エラー発生時の処理
        indicator.textContent = "アップロード中にエラーが発生しました: " + err.message;
        console.error(err);
      }
    }
    // 全ファイル処理完了後のメッセージ
    indicator.textContent += " 完了";
    // 入力値のリセット
    input.value = '';
  });
});