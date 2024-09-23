paypal.Buttons({
    style: {
        shape: "pill",
        layout: "vertical",
        color: "black",
        label: "pay"
    },
    async createOrder() {
        const savedData = JSON.parse(localStorage.getItem('conex_data')) || {};
        const requestData = {
            directory: savedData.directory,
        };
        const response = await fetch("https://api.conex.one/api/orders", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify(requestData),
        });

        if (!response.ok) {
            if (response.status === 409) {
                checkoutError(`
                    <p>El título "${savedData.title}" es incorrecto, debe cumplir:<br>
                    <ul>
                    <li>Entre 4 y 35 caracteres</li>
                    <li>Debe ser único</li>
                    </ul>
                    </p>
                `);
            } else {
                checkoutError(`<p>No se puede realizar la compra en este momento</p>`);
            }
            console.log(`HTTP Error: ${response.status} - ${response.statusText}`);
            return;
        }

        const orderData = await response.json();

        if (orderData.id) {
            return orderData.id;
        } else {
            const errorDetail = orderData?.details?.[0];
            checkoutError(`<p>No se puede realizar la compra en este momento</p>`);
        }
    },
    async onApprove(data, actions) {
        const savedData = JSON.parse(localStorage.getItem('conex_data')) || {};
        try {
            const requestData = {
                directory: savedData.directory,
                banner: savedData.banner,
                title: savedData.title,
                slogan: savedData.slogan,
                tags: savedData.tags,
                editor_data: savedData.editor_data
            };

            const response = await fetch(`https://api.conex.one/api/orders/${data.orderID}/capture`, {
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
                checkoutSuccess(`
                        <p>
                        Estado: <strong>${transaction.status}</strong><br>
                        ID de transacción: ${transaction.id}<br>
                        Luego de una revisión positiva, su sitio será publicado en menos de 24 horas en el enlace:
                        </p>
                        <p>
                        <a href="https://conex.one/${savedData.directory}/">conex.one/${savedData.directory}</a>
                        </p>
                `,);
                document.querySelector("#paypal-button-container").style.display = "none";
                console.log(
                    "Capture result",
                    orderData,
                    JSON.stringify(orderData, null, 2),
                );
            }
        } catch (error) {
            console.error(error);
            checkoutError(`<p>No se puede realizar la compra en este momento</p>`);
        }
    },
}).render("#paypal-button-container");

function checkoutSuccess(message) {
    const container = document.querySelector("#checkout-success-message");
    container.style.display = "block";
    container.innerHTML = message;
}

function checkoutError(message) {
    const container = document.querySelector("#checkout-error-message");
    container.style.display = "block";
    container.innerHTML = message;
}
