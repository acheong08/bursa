class AlertsEditor extends HTMLElement {
  constructor() {
    super();
    this.attachShadow({ mode: "open" });
    this.alerts = [];
    this.variables = [
      "last_price",
      "preclose_price",
      "price_change",
      "total_bought_quantity",
      "trade_value",
      "buy_volume",
      "sell_volume",
      "buy_rate",
    ];
  }

  async connectedCallback() {
    this.render();
    this.addEventListeners();
    await this.fetchAlerts();
    this.renderAlertsList();
  }

  render() {
    this.shadowRoot.innerHTML = `
      <style>
        /* Styles for the main component */
        :host {
          display: block;
          font-family: Arial, sans-serif;
        }
        #alerts-list {
          margin-bottom: 10px;
        }
        /* Styles for the dialog */
        dialog {
          padding: 20px;
          border: 1px solid #ccc;
          border-radius: 5px;
          max-width: 500px;
        }
        form {
          display: flex;
          flex-direction: column;
        }
        label {
          margin-top: 10px;
        }
        input, select {
          margin-bottom: 10px;
        }
        .rule {
          border: 1px solid #eee;
          padding: 10px;
          margin-bottom: 10px;
        }
        .tags {
          display: flex;
          flex-wrap: wrap;
          gap: 5px;
        }
        .tag {
          background-color: #eee;
          border-radius: 3px;
          padding: 2px 5px;
        }
      </style>
			<h3>Alerts</h3>
      <div id="alerts-list"></div>
      <button id="add-alert">Add Alert</button>
      <dialog id="alert-dialog"></dialog>
    `;
  }

  addEventListeners() {
    this.shadowRoot
      .querySelector("#add-alert")
      .addEventListener("click", () => this.openAlertDialog());
  }

  async fetchAlerts() {
    // Fetch alerts from the server and update this.alerts
    // Render the list of alerts in the main view
    const resp = await fetch("/alerts/");
    if (resp.status !== 200) {
      alert(await resp.text());
      return;
    }
    this.alerts = await resp.json();
    if (this.alerts === null) {
      this.alerts = [];
    }
    return;
  }

  openAlertDialog(index) {
    const dialog = this.shadowRoot.querySelector("#alert-dialog");
    const alert = index !== undefined ? this.alerts[index] : null;

    dialog.innerHTML = `
      <form id="alert-form">
        <label for="label">Label:</label>
        <input type="text" id="label" name="label" value="${alert ? alert.label : ""}" required>
        
        <label for="tags">Tags (comma-separated):</label>
        <input type="text" id="tags" name="tags" value="${alert ? (alert.tags ? alert.tags.join(",") : "") : ""}">
        
        <h3>Rules:</h3>
        <div id="rules-container">
          ${alert ? this.renderRules(alert.rules) : ""}
        </div>
        <button type="button" id="add-rule">Add Rule</button>
        
        <button type="submit">Save</button>
        <button type="button" id="cancel">Cancel</button>
      </form>
    `;

    dialog.querySelector("#alert-form").addEventListener("submit", (e) => {
      e.preventDefault();
      this.saveAlert(index);
    });

    dialog
      .querySelector("#cancel")
      .addEventListener("click", () => dialog.close());
    dialog
      .querySelector("#add-rule")
      .addEventListener("click", () => this.addRule());
    dialog.addEventListener("change", (e) => {
      if (e.target.classList.contains("type-select")) {
        const { index, side } = e.target.dataset;
        const valueContainer = e.target.nextElementSibling;
        const newInput = this.renderValueInput(side, index, {
          type: e.target.value,
          value: "",
        });
        valueContainer.outerHTML = newInput;
      }
    });
    dialog.showModal();
  }

  renderRules(rules) {
    if (!rules) return "";
    return rules
      .map(
        (rule, index) => `
    <div class="rule">
      <select name="a-type-${index}" class="type-select" data-index="${index}" data-side="a">
        <option value="var" ${rule.a.type === "var" ? "selected" : ""}>Variable</option>
        <option value="const" ${rule.a.type === "const" ? "selected" : ""}>Constant</option>
      </select>
      ${this.renderValueInput("a", index, rule.a)}
      
      <select name="cmp-${index}">
        <option value="==" ${rule.cmp === "==" ? "selected" : ""}>==</option>
        <option value="!=" ${rule.cmp === "!=" ? "selected" : ""}>!=</option>
        <option value=">" ${rule.cmp === ">" ? "selected" : ""}>></option>
        <option value=">=" ${rule.cmp === ">=" ? "selected" : ""}>>=</option>
        <option value="<" ${rule.cmp === "<" ? "selected" : ""}><</option>
        <option value="<=" ${rule.cmp === "<=" ? "selected" : ""}><=</option>
      </select>
      
      <select name="b-type-${index}" class="type-select" data-index="${index}" data-side="b">
        <option value="var" ${rule.b.type === "var" ? "selected" : ""}>Variable</option>
        <option value="const" ${rule.b.type === "const" ? "selected" : ""}>Constant</option>
      </select>
      ${this.renderValueInput("b", index, rule.b)}
      
      <button type="button" class="remove-rule" data-index="${index}">Remove</button>
    </div>
  `,
      )
      .join("");
  }

  renderValueInput(side, index, data) {
    if (data.type === "var") {
      return `
      <select name="${side}-value-${index}">
        ${this.variables.map((v) => `<option value="${v}" ${data.value === v ? "selected" : ""}>${v}</option>`).join("")}
      </select>
    `;
    }
    return `<input type="text" name="${side}-value-${index}" value="${data.value}" required>`;
  }
  addRule() {
    const rulesContainer = this.shadowRoot.querySelector("#rules-container");
    const newRuleIndex = rulesContainer.children.length;
    const newRuleHtml = this.renderRules([
      {
        a: { type: "var", value: "" },
        cmp: "==",
        b: { type: "const", value: "" },
      },
    ]);
    rulesContainer.insertAdjacentHTML("beforeend", newRuleHtml);

    // Add event listener for the new remove button
    const newRemoveButton =
      rulesContainer.lastElementChild.querySelector(".remove-rule");
    newRemoveButton.addEventListener("click", () =>
      this.removeRule(newRuleIndex),
    );
  }

  removeRule(index) {
    const rulesContainer = this.shadowRoot.querySelector("#rules-container");
    rulesContainer.children[index].remove();
  }

  async saveAlert(index) {
    const form = this.shadowRoot.querySelector("#alert-form");
    const formData = new FormData(form);

    const newAlert = {
      label: formData.get("label"),
      tags: formData
        .get("tags")
        .split(",")
        .map((tag) => tag.trim())
        .filter((tag) => tag),
      rules: [],
    };

    // Collect rules
    const ruleElements = this.shadowRoot.querySelectorAll(".rule");
    ruleElements.forEach((_, i) => {
      newAlert.rules.push({
        a: {
          type: formData.get(`a-type-${i}`),
          value: formData.get(`a-value-${i}`),
        },
        cmp: formData.get(`cmp-${i}`),
        b: {
          type: formData.get(`b-type-${i}`),
          value: formData.get(`b-value-${i}`),
        },
      });
    });

    // Here you would typically send this data to your server
    const resp = await fetch("/alerts/" + (index ? "?id=" + index : ""), {
      method: index === undefined ? "PUT" : "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(newAlert),
    });
    if (resp.status !== 200) {
      alert(await resp.text());
      return;
    }
    if (index === undefined) {
      this.alerts.push(newAlert);
    }

    // Close the dialog and update the view
    this.shadowRoot.querySelector("#alert-dialog").close();
    this.renderAlertsList();
  }

  async deleteAlert(index) {
    const resp = await fetch(`/alerts/?id=${index}`, {
      method: "DELETE",
    });
    if (resp.status !== 200) {
      alert(await resp.text());
      return;
    }

    this.alerts.splice(index, 1);
    this.renderAlertsList();
  }

  renderAlertsList() {
    const alertsList = this.shadowRoot.querySelector("#alerts-list");
    alertsList.innerHTML = this.alerts
      .map(
        (alert, index) => `
			<div class="alert">
				<h4>${alert.label}</h4>
				<div class="tags">
					${alert.tags ? alert.tags.map((tag) => `<span class="tag">${tag}</span>`).join("") : ""}
				</div>
				<button type="button" class="edit-alert" data-index="${index}">Edit</button>
				<button type="button" class="delete-alert" data-index="${index}">Delete</button>
			</div>
		`,
      )
      .join("");

    alertsList.querySelectorAll(".edit-alert").forEach((button) => {
      button.addEventListener("click", () =>
        this.openAlertDialog(button.dataset.index),
      );
    });

    alertsList.querySelectorAll(".delete-alert").forEach((button) => {
      button.addEventListener("click", () =>
        this.deleteAlert(button.dataset.index),
      );
    });
  }
}

customElements.define("alerts-editor", AlertsEditor);
