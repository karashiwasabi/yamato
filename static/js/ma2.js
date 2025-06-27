// static/js/ma2.js

(() => {
  // ── 1) サーバーから単位マップを取得 ──
  let taniMap = {};
  fetch("/api/tani")
    .then(res => res.json())
    .then(map => { taniMap = map; })
    .catch(err => console.error("TANI取得失敗:", err));

  document.addEventListener("DOMContentLoaded", () => {
    const ma2Btn     = document.getElementById("ma2Btn");
    const editor     = document.querySelector(".ma2-editor");
    const headerGrid = document.querySelector(".ma2-grid.header");
    const bodyGrid   = document.querySelector(".ma2-grid.body");

    // 管理するフィールドキーと見出し
    const COLUMNS = [
      "janCode",
      "yjCode",
      "shouhinmei",
      "housouKeitai",
      "housouTaniUnit",
      "housouSouryouNumber",
      "janHousouSuuryouNumber",
      "janHousouSuuryouUnit",
      "janHousouSouryouNumber"
    ];
    const HEADERS = [
      "JANコード",
      "YJコード",
      "商品名",
      "包装形態",
      "包装単位",
      "包装総量",
      "JAN包装数量",
      "JAN包装数量単位",
      "JAN包装総量",
      "操作"
    ];

    // ── ヘッダー行を描画 ──
    function renderHeader() {
      headerGrid.innerHTML = "";
      HEADERS.forEach(text => {
        const cell = document.createElement("div");
        cell.textContent = text;
        cell.className = "header-cell";
        headerGrid.appendChild(cell);
      });
    }

    // ── 1レコード分の行を作る ──
    function makeRow(data = {}, isNew = false) {
      const row = document.createElement("div");
      row.className = "row";

      COLUMNS.forEach(key => {
        // セル用コンテナ（枠線などは CSS 側で制御）
        let cell;
        if (key === "housouTaniUnit" || key === "janHousouSuuryouUnit") {
          // select プルダウン
          const sel = document.createElement("select");
          const empty = document.createElement("option");
          empty.value = "";
          sel.appendChild(empty);
          Object.entries(taniMap).forEach(([code, name]) => {
            const opt = document.createElement("option");
            opt.value = code;
            opt.textContent = name;
            if (data[key] === code) opt.selected = true;
            sel.appendChild(opt);
          });
          cell = sel;
        } else {
          // input テキスト/数値
          const inp = document.createElement("input");
          // 数値系は type="number"
          if (["housouSouryouNumber","janHousouSuuryouNumber","janHousouSouryouNumber"]
              .includes(key)) {
            inp.type = "number";
            inp.min  = "0";
          } else {
            inp.type = "text";
          }
          // 既存レコードの JANコードは編集不可
          if (!isNew && key === "janCode") {
            inp.readOnly = true;
            inp.classList.add("readonly");
          }
          inp.value = data[key] != null ? data[key] : "";
          cell = inp;
        }

        row.appendChild(cell);
      });

      // ── 操作ボタン ──
      const btn = document.createElement("button");
      btn.textContent = isNew ? "登録" : "更新";
      btn.className = "btn";
      btn.addEventListener("click", async () => {
        // 送信レコードを組み立て
        const rec = {};
        COLUMNS.forEach((key, idx) => {
          const el = row.children[idx];
          let v = (el.tagName === "SELECT")
                  ? el.value
                  : el.value.trim();
          // 数値は parseInt
          if (["housouSouryouNumber","janHousouSuuryouNumber","janHousouSouryouNumber"]
              .includes(key)) {
            v = parseInt(v, 10) || 0;
          }
          rec[key] = v;
        });

        // upsert リクエスト
        try {
          const res = await fetch("/api/ma2/upsert", {
            method:  "POST",
            headers: { "Content-Type": "application/json" },
            body:    JSON.stringify(rec)
          });
          if (!res.ok) throw new Error(res.statusText);
          alert((isNew ? "登録" : "更新") + " 成功");
          loadData();
        } catch (e) {
          alert("エラー: " + e.message);
        }
      });
      row.appendChild(btn);

      return row;
    }

    // ── データ取得＆描画 ──
    async function loadData() {
      // outputTable は common.js が隠すので、ここでは editor 表示
      editor.style.display = "block";

      // グリッドをクリア
      renderHeader();
      bodyGrid.innerHTML = "";

      // 新規登録用行
      bodyGrid.appendChild(makeRow({}, true));

      // 既存データ
      try {
        const res  = await fetch("/api/ma2");
        if (!res.ok) throw new Error(res.statusText);
        const data = await res.json();
        data.forEach(rec => bodyGrid.appendChild(makeRow(rec, false)));
      } catch (e) {
        alert("データ取得失敗: " + e.message);
      }
    }

    ma2Btn.addEventListener("click", loadData);
  });
})();