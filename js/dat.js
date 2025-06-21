document.addEventListener('DOMContentLoaded', () => {
  const uploadBtn = document.getElementById('uploadBtn');
  const fileInput = document.getElementById('fileInput');
  const indicator = document.getElementById('indicator');
  const output = document.getElementById('output');

  if (!uploadBtn || !fileInput || !indicator || !output) {
    console.error("必要な DOM 要素が見つかりません。HTML を確認してください。");
    return;
  }

  // アップロードボタンをクリックでファイル選択ダイアログを表示
  uploadBtn.addEventListener('click', () => {
    fileInput.click();
  });

  fileInput.addEventListener('change', async () => {
    const files = fileInput.files;
    if (!files || files.length === 0) {
      console.warn("ファイルが選択されていません。");
      return;
    }

    // Indicator に選択されたファイル数を表示
    indicator.textContent = `${files.length} 個のファイルが選択されました。`;

    let results = [];
    // 各ファイルを1個ずつアップロード
    for (let i = 0; i < files.length; i++) {
      const formData = new FormData();
      formData.append('file', files[i]);

      try {
        const res = await fetch('/upload', {
          method: 'POST',
          body: formData
        });
        if (!res.ok) throw new Error(`HTTPエラー: ${res.status}`);
        // レスポンスはプレーンテキストとして受け取る
        const text = await res.text();
        results.push(`[${files[i].name}]:\n${text}`);
      } catch (err) {
        results.push(`[${files[i].name}]: アップロード失敗 (${err.message})`);
      }
    }
    output.textContent = results.join("\n\n");
    indicator.textContent += " アップロード処理が完了しました。";
    fileInput.value = '';
  });
});