// static/js/common.js
document.addEventListener("DOMContentLoaded", () => {
  const table      = document.getElementById("outputTable");
  const thead      = table.querySelector("thead");
  const tbody      = table.querySelector("tbody");
  const filter     = document.getElementById("aggregateFilter");
  const editor     = document.querySelector(".ma2-editor");
  const indicator  = document.getElementById("indicator");
  const debug      = document.getElementById("debug");
  const inoutForm  = document.getElementById("inoutForm");  // ← 追加

  function resetUI() {
    // MA2エディタを隠す
    editor.style.display      = "none";
    // 集計／DAT／USAGE／棚卸テーブルを表示
    table.style.display       = "table";
    // フィルタ部を隠す
    filter.style.display      = "none";
    // 出庫／入庫フォームを隠す
    inoutForm.classList.add("hidden");
    // テーブルクリア
    thead.innerHTML           = "";
    tbody.innerHTML           = "";
    // インジケータ・デバッグ領域クリア
    indicator.textContent     = "";
    debug.textContent         = "";
  }

  // NAV の全ボタンで resetUI を実行（出庫・入庫ボタンも含む）
  document
    .querySelectorAll("header nav .btn")
    .forEach(btn => btn.addEventListener("click", resetUI, true));
});