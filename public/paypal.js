paypal.Buttons({
  style: {
    shape: "pill",
    layout: "vertical",
    color: "black",
    label: "pay"
  },
  async createOrder() {
    const savedData = JSON.parse(localStorage.getItem('editor_data')) || {};
    const requestData = {
      directory: sanitizeDirectoryName(savedData.title),
      banner: savedData.banner || '/static/svg/banner.svg',
      title: savedData.title,
      slogan: savedData.slogan,
      editor_data: savedData.editor_data
    };
    const response = await fetch("/api/orders", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(requestData),
    });

    if (!response.ok) {
      if (response.status === 409) {
        resultMessage(`No se puede comprar este sitio, ya existe o tiene un nombre incorrecto. Prueba con un nombre diferente`);
      } else {
        resultMessage(`No se puede realizar la compra en este momento`);
      }
      console.log(`HTTP Error: ${response.status} - ${response.statusText}`);
      return;
    }

    const orderData = await response.json();

    if (orderData.id) {
      return orderData.id;
    } else {
      const errorDetail = orderData?.details?.[0];
      resultMessage(`No se puede realizar la compra en este momento`);
    }
  },
  async onApprove(data, actions) {
    try {
      // @@@ TODO por alguna razon hizo la compra a pesar de que esto estaba mal puesto
      const requestData = {
        directory: "gofitness",
        editor_data: await editor.save()
      };

      const response = await fetch(`/api/orders/${data.orderID}/capture`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(requestData),
      });

      const orderData = await response.json();
      // Three cases to handle:
      //   (1) Recoverable INSTRUMENT_DECLINED -> call actions.restart()
      //   (2) Other non-recoverable errors -> Show a failure message
      //   (3) Successful transaction -> Show confirmation or thank you message

      const errorDetail = orderData?.details?.[0];

      if (errorDetail?.issue === "INSTRUMENT_DECLINED") {
        // (1) Recoverable INSTRUMENT_DECLINED -> call actions.restart()
        // recoverable state, per https://developer.paypal.com/docs/checkout/standard/customize/handle-funding-failures/
        return actions.restart();
      } else if (errorDetail) {
        // (2) Other non-recoverable errors -> Show a failure message
        throw new Error(`${errorDetail.description} (${orderData.debug_id})`);
      } else if (!orderData.purchase_units) {
        throw new Error(JSON.stringify(orderData));
      } else {
        // (3) Successful transaction -> Show confirmation or thank you message
        // Or go to another URL:  actions.redirect('thank_you.html');
        const transaction =
          orderData?.purchase_units?.[0]?.payments?.captures?.[0] ||
          orderData?.purchase_units?.[0]?.payments?.authorizations?.[0];
        resultMessage(
          `Estado: <strong>${transaction.status}</strong><br>ID de transacción: ${transaction.id}<br>Luego de una revisión positiva, su sitio será publicado en menos de 24 horas.`,
        );
        console.log(
          "Capture result",
          orderData,
          JSON.stringify(orderData, null, 2),
        );
      }
    } catch (error) {
      console.error(error);
      resultMessage(
        `Sorry, your transaction could not be processed...<br><br>${error}`,
      );
    }
  },
}).render("#paypal-button-container");

// Example function to show a result to the user. Your site's UI library can be used instead.
function resultMessage(message) {
  const container = document.querySelector("#result-message");
  container.innerHTML = message;
}
