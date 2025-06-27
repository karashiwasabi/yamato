// static/js/ma2.js

(() => {
  // 1) サーバーから TANI マップを取得
  let taniMap = {};
  fetch("/api/tani")
    .then(res => res.json())
    .then(map => { taniMap = map; })
    .catch(err => console.error("TANI取得失敗:", err));

  document.addEventListener("DOMContentLoaded", () => {
    const ma2Btn      = document.getElementById("ma2Btn");
    const outputTable = document.getElementById("outputTable");
    const editor      = document.querySelector(".ma2-editor");
    const headerThead = document.querySelector("#ma2Header thead");
    const bodyTbody   = document.querySelector("#ma2Body tbody");

    // 管理する列キーと見出し
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

    // ヘッダーを描画
    function renderHeader() {
      headerThead.innerHTML = "";
      const tr = document.createElement("tr");
      HEADERS.forEach(h => {
        const th = document.createElement("th");
        th.textContent = h;
        tr.appendChild(th);
      });
      headerThead.appendChild(tr);
    }

    // 1行分の<tr>要素を生成
    function makeRow(data = {}, isNew = false) {
      const tr = document.createElement("tr");

      COLUMNS.forEach((key, idx) => {
        const td = document.createElement("td");

        // 包装単位系は <select> に
        if (key === "housouTaniUnit" || key === "janHousouSuuryouUnit") {
          const sel = document.createElement("select");
          sel.style.minWidth = "4em";
          const empty = document.createElement("option");
          empty.value = "";
          empty.textContent = "";
          sel.appendChild(empty);
          Object.entries(taniMap).forEach(([code, name]) => {
            const opt = document.createElement("option");
            opt.value = code;
            opt.textContent = name;
            if (data[key] === code) opt.selected = true;
            sel.appendChild(opt);
          });
          td.appendChild(sel);

        } else if (key === "janCode" && !isNew) {
          // 既存レコードの JAN は編集不可
          td.textContent = data[key] || "";

        } else {
          // その他は contentEditable
          td.contentEditable = true;
          td.textContent = data[key] != null ? data[key] : "";
        }

        tr.appendChild(td);
      });

      // 操作セル：登録 or 更新
      const opTd = document.createElement("td");
      const opBtn = document.createElement("button");
      opBtn.className = "btn";
      opBtn.textContent = isNew ? "登録" : "更新";
      opBtn.addEventListener("click", async () => {
        // 入力値をまとめる
        const rec = {};
        COLUMNS.forEach((key, idx) => {
          let v;
          if (key === "housouTaniUnit" || key === "janHousouSuuryouUnit") {
            v = tr.children[idx].querySelector("select").value;
          } else {
            v = tr.children[idx].textContent.trim();
            if (["housouSouryouNumber","janHousouSuuryouNumber","janHousouSouryouNumber"]
                .includes(key)) {
              v = parseInt(v, 10) || 0;
            }
          }
          rec[key] = v;
        });

        // サーバーに upsert リクエスト
        try {
          const res = await fetch("/api/ma2/upsert", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(rec)
          });
          if (!res.ok) throw new Error(res.statusText);
          alert((isNew ? "登録" : "更新") + " 成功");
          loadData();
        } catch (e) {
          alert("エラー: " + e.message);
        }
      });
      opTd.appendChild(opBtn);
      tr.appendChild(opTd);

      return tr;
    }

    // データ取得→描画
    async function loadData() {
      // 他機能テーブルを隠して
      outputTable.style.display = "none";
      // MA2エディタを表示
      editor.style.display = "block";

      renderHeader();
      bodyTbody.innerHTML = "";

      // 新規登録用の空行
      bodyTbody.appendChild(makeRow({}, true));

      try {
        const res = await fetch("/api/ma2");
        if (!res.ok) throw new Error(res.statusText);
        const data = await res.json();
        data.forEach(d => bodyTbody.appendChild(makeRow(d, false)));
      } catch (err) {
        alert("データ取得失敗: " + err.message);
      }
    }

    ma2Btn.addEventListener("click", loadData);
  });
})();