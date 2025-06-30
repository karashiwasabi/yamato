// static/js/ma2.js
(() => {
  // ── 1) 包装単位マップ取得 ──
  let taniMap = {};
  fetch("/api/tani")
    .then(res => res.json())
    .then(map => { taniMap = map; })
    .catch(err => console.error("TANI取得失敗:", err));

  document.addEventListener("DOMContentLoaded", () => {
    const ma2Btn   = document.getElementById("ma2Btn");
    const editor   = document.querySelector(".ma2-editor");
    const bodyGrid = document.querySelector(".ma2-grid.body");

    // ── フィールド定義 ──
    const FIELDS = [
      { key: "janCode",                label: "JANコード",           type: "text",    readOnly: true  },
      { key: "yjCode",                 label: "YJコード",            type: "text"                   },
      { key: "shouhinmei",             label: "商品名",              type: "text"                   },
      { key: "housouKeitai",           label: "包装形態",            type: "text"                   },
      { key: "housouSouryouNumber",    label: "包装総量",            type: "number"                 },
      { key: "housouTaniUnit",         label: "包装単位",            type: "select"                 },
      { key: "janHousouSuuryouNumber", label: "JAN包装数量",         type: "number"                 },
      { key: "janHousouSouryouNumber", label: "JAN包装総量",         type: "number"                 },
      { key: "janHousouSuuryouUnit",   label: "JAN包装数量単位",     type: "select"                 }
    ];

    ma2Btn.addEventListener("click", loadData);

    // ── フィールド要素（input/select）を生成 ──
    function createFieldCell(key, value = "", isNew) {
      const def = FIELDS.find(f => f.key === key);
      let el;
      if (def.type === "select") {
        el = document.createElement("select");
        el.innerHTML = `<option value=""></option>`;
        Object.entries(taniMap).forEach(([code, name]) => {
          const o = document.createElement("option");
          o.value = code;
          o.textContent = name;
          if (value === code) o.selected = true;
          el.appendChild(o);
        });
      } else {
        el = document.createElement("input");
        el.type = def.type;
        if (def.type === "number") el.min = "0";
        if (!isNew && def.readOnly) el.readOnly = true;
        el.value = value;
      }
      return el;
    }

    // ── １レコードを<table>で生成（<thead>ラベル行・<tbody>入力行・<tfoot>操作行＋プレビュー） ──
    function makeRecord(data = {}, isNew = false) {
      const table = document.createElement("table");
      table.className = "record";

      // thead：ラベル行
      const thead = document.createElement("thead");
      const headTr = document.createElement("tr");
      FIELDS.forEach(def => {
        const th = document.createElement("th");
        th.textContent = def.label;
        headTr.appendChild(th);
      });
      thead.appendChild(headTr);
      table.appendChild(thead);

      // tbody：入力行
      const tbody = document.createElement("tbody");
      const bodyTr = document.createElement("tr");
      FIELDS.forEach(def => {
        const td = document.createElement("td");
        const val = data[def.key] != null ? data[def.key] : "";
        td.appendChild(createFieldCell(def.key, val, isNew));
        bodyTr.appendChild(td);
      });
      tbody.appendChild(bodyTr);
      table.appendChild(tbody);

      // tfoot：操作行＋プレビュー
      const tfoot = document.createElement("tfoot");
      const footTr = document.createElement("tr");
      const footTd = document.createElement("td");
      footTd.colSpan = FIELDS.length;

      // 登録／更新ボタン
      const btn = document.createElement("button");
      btn.textContent = isNew ? "登録" : "更新";
      btn.addEventListener("click", () => upsertRecord(table, isNew));
      footTd.appendChild(btn);

      // プレビュー用 span
      const preview = document.createElement("span");
      preview.className = "labelPreview";
      preview.style.marginLeft = "8px";
      footTd.appendChild(preview);

      footTr.appendChild(footTd);
      tfoot.appendChild(footTr);
      table.appendChild(tfoot);

      // ── プレビュー更新ロジック ──
      const inputs = {};
      FIELDS.forEach((def, i) => {
        const cell = bodyTr.children[i];
        inputs[def.key] = cell.querySelector(def.type === "select" ? "select" : "input");
      });
      function updatePreview() {
        const f   = inputs.housouKeitai.value.trim();
        const s   = inputs.housouSouryouNumber.value.trim();
        const ut1 = inputs.housouTaniUnit.selectedOptions[0]?.text || "";
        const jn  = inputs.janHousouSuuryouNumber.value.trim();
        const mn  = inputs.janHousouSouryouNumber.value.trim();
        const ut2 = inputs.janHousouSuuryouUnit.selectedOptions[0]?.text || "";
        preview.textContent = `${f}${s}${ut1}（${jn}${ut1}×${mn}${ut2}）`;
      }
      // イベント紐付け
      ["housouKeitai","housouSouryouNumber","housouTaniUnit",
       "janHousouSuuryouNumber","janHousouSouryouNumber","janHousouSuuryouUnit"
      ].forEach(key => {
        const el = inputs[key];
        const ev = el.tagName === "SELECT" ? "change" : "input";
        el.addEventListener(ev, updatePreview);
      });
      updatePreview();

      return table;
    }

    // ── データ取得＆描画 ──
    async function loadData() {
      editor.style.display = "block";
      bodyGrid.innerHTML   = "";

      // 新規入力行
      bodyGrid.appendChild(makeRecord({}, true));

      // 既存データ行
      try {
        const res  = await fetch("/api/ma2");
        if (!res.ok) throw new Error(res.statusText);
        const data = await res.json();
        data.forEach(rec => bodyGrid.appendChild(makeRecord(rec, false)));
      } catch (e) {
        alert("データ取得失敗: " + e.message);
      }
    }

    // ── upsert API 呼び出し ──
    async function upsertRecord(tableEl, isNew) {
      const rec    = {};
      const elems  = tableEl.querySelectorAll("input,select");
      elems.forEach((el, i) => {
        const key = FIELDS[i].key;
        let v = el.tagName === "SELECT" ? el.value : el.value.trim();
        if (FIELDS[i].type === "number") v = parseInt(v, 10) || 0;
        rec[key] = v;
      });

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
    }
  });
})();