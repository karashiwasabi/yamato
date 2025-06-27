// static/js/common.js
document.addEventListener("DOMContentLoaded", () => {
  const table   = document.getElementById("outputTable");
  const thead   = table.querySelector("thead");
  const tbody   = table.querySelector("tbody");
  const filter  = document.getElementById("aggregateFilter");
  const editor  = document.querySelector(".ma2-editor");
  const indicator = document.getElementById("indicator");
  const debug   = document.getElementById("debug");

  function resetUI() {
    // テーブル画面に戻す
    editor.style.display      = "none";
    table.style.display       = "table";
    filter.style.display      = "none";
    // 各表示をクリア
    thead.innerHTML           = "";
    tbody.innerHTML           = "";
    indicator.textContent     = "";
    debug.textContent         = "";
  }

  // NAVのどのボタンでも先にUIリセット
  document
    .querySelectorAll("header nav .btn")
    .forEach(btn => btn.addEventListener("click", resetUI, true));
});