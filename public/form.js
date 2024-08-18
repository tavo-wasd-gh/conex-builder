// TODO
//
// 1. Try to disable asking for shipping info (although could be
//    useful to mark as sent).
//
// 2. Read about IPN and Webhooks to automate registering process.

const clientId = "";
const OneTimePID = "";
const PlanID = "";

const form = document.getElementById('mainForm');

['name', 'email', 'phone'].forEach(id => {
    const input = document.createElement('input');
    input.type = 'hidden';
    input.name = id;
    input.value = document.getElementById(id).value;
    form.appendChild(input);
});

function hideDialog() {
    document.getElementById('overlay').style.display = 'none';
    document.getElementById('dialog').style.display = 'none';
    document.getElementById('openDialogButton').style.display = 'block';
}

function showDialog() {
    document.getElementById('overlay').style.display = 'block';
    document.getElementById('dialog').style.display = 'block';
    document.getElementById('openDialogButton').style.display = 'none';
}

function togglePaymentMethod(selectedButtonId) {
    // Deselect all buttons and hide all PayPal buttons
    document.querySelectorAll('#method-button-container button').forEach(button => {
        button.classList.remove('active');
    });
    document.querySelectorAll('#paypal-button-container > div').forEach(div => {
        div.classList.remove('active');
    });

    // Select the clicked button and show the corresponding PayPal button
    const selectedButton = document.getElementById(selectedButtonId);
    selectedButton.classList.add('active');

    if (selectedButtonId === 'showOneTimeButton') {
        document.getElementById('paypal-button-container').classList.add('active');
        document.getElementById('paypal-button-container-order').classList.add('active');
    } else if (selectedButtonId === 'showSubButton') {
        document.getElementById('paypal-button-container').classList.add('active');
        document.getElementById('paypal-button-container-subscribe').classList.add('active');
    }
}

function isFormValid(form) {
  return form.checkValidity();
}

window.paypal_order.Buttons({
    style: { shape: 'pill', color: 'black', layout: 'vertical', label: 'pay' },
    async createOrder() {
        try {
            const response = await fetch("/api/orders", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                // use the "body" param to optionally pass additional order information
                // like product ids and quantities
                body: JSON.stringify({
                    cart: [
                        {
                            id: "YOUR_PRODUCT_ID",
                            quantity: "YOUR_PRODUCT_QUANTITY",
                        },
                    ],
                }),
            });

            const orderData = await response.json();

            if (orderData.id) {
                return orderData.id;
            } else {
                const errorDetail = orderData?.details?.[0];
                const errorMessage = errorDetail
                    ? `${errorDetail.issue} ${errorDetail.description} (${orderData.debug_id})`
                    : JSON.stringify(orderData);

                throw new Error(errorMessage);
            }
        } catch (error) {
            console.error(error);
            resultMessage(`Could not initiate PayPal Checkout...<br><br>${error}`);
        }
    },
    async onApprove(data, actions) {
        try {
            const response = await fetch(`/api/orders/${data.orderID}/capture`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
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
                    `Transaction ${transaction.status}: ${transaction.id}<br><br>See console for all available details`,
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
}).render("#paypal-button-container-order");

// Example function to show a result to the user. Your site's UI library can be used instead.
function resultMessage(message) {
  const container = document.querySelector("#checkout");
  container.innerHTML = message;
}

document.getElementById('showOneTimeButton').addEventListener('click', function() {
  if (isFormValid(form)) {
    document.getElementById('warning-message').style.display = 'none';
    togglePaymentMethod('showOneTimeButton');
  } else {
    document.getElementById('warning-message').style.display = 'block';
  }
});

document.getElementById('showSubButton').addEventListener('click', function() {
  if (isFormValid(form)) {
    document.getElementById('warning-message').style.display = 'none';
    togglePaymentMethod('showSubButton');
  } else {
    document.getElementById('warning-message').style.display = 'block';
  }
});

document.getElementById('openDialogButton').addEventListener('click', () => {
    showDialog();
});

document.getElementById('cancelDialogButton').addEventListener('click', () => {
    hideDialog();
});
