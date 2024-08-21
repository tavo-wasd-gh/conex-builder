document.addEventListener("DOMContentLoaded", function() {
    const dialog = document.getElementById("dialog");
    const overlay = document.getElementById("overlay");
    const menu = document.getElementById("floatingButtons");

    function openDialog() {
        dialog.style.display = "block";
        overlay.style.display = "block";
        menu.style.display = "none";
    }

    function closeDialog() {
        dialog.style.display = "none";
        overlay.style.display = "none";
        menu.style.display = "block";
    }

    function togglePaymentMethod(selectedButtonId) {
        // Deselect all buttons and hide all PayPal buttons
        document.querySelectorAll('#method-button-container button').forEach(button => { button.classList.remove('active'); });
        document.querySelectorAll('#paypal-button-container > div').forEach(div => { div.classList.remove('active'); });

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

    document.getElementById('showOneTimeButton').addEventListener('click', function() {
        document.getElementById('warning-message').style.display = 'none';
        togglePaymentMethod('showOneTimeButton');
    });

    document.getElementById('showSubButton').addEventListener('click', function() {
        document.getElementById('warning-message').style.display = 'none';
        togglePaymentMethod('showSubButton');
    });

    document.getElementById("openDialogButton").addEventListener("click", openDialog);
    document.getElementById("cancelDialogButton").addEventListener("click", closeDialog);
});


window.paypal_order.Buttons({
    style: { shape: 'pill', color: 'black', layout: 'vertical', label: 'pay' },
    async createOrder() {
        try {
            const response = await fetch("/api/order", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
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
            const response = await fetch(`/api/order/${data.orderID}/capture`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify(
                    {
                        directory: "tutorias",
                    }
                ),
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

window.paypal_subscribe.Buttons({
    style: { shape: 'pill', color: 'black', layout: 'vertical', label: 'subscribe' },
    async createSubscription() {
        try {
            const response = await fetch("/api/paypal/subscribe", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify(
                  {
                    // userAction: "SUBSCRIBE_NOW"
                    directory: "testsite",
                  }
                ),
            });
            const data = await response.json();
            if (data?.id) {
                const approvalUrl = data.links.find(link => link.rel === "approve").href;
                window.location.href = approvalUrl;
                resultMessage(`Successful subscription with ID ${approvalUrl}...<br><br>`);
                // resultMessage(`Successful subscription with ID ${data.id}...<br><br>`);
                return data.id;
            } else {
                console.error(
                    { callback: "createSubscription", serverResponse: data },
                    JSON.stringify(data, null, 2),
                );
                // (Optional) The following hides the button container and shows a message about why checkout can't be initiated
                const errorDetail = data?.details?.[0];
                resultMessage(
                    `Could not initiate PayPal Subscription...<br><br>${
                        errorDetail?.issue || ""
                    } ${errorDetail?.description || data?.message || ""} ` +
                    (data?.debug_id ? `(${data.debug_id})` : ""),
                    { hideButtons: true },
                );
            }
        } catch (error) {
            console.error(error);
            resultMessage(
                `Could not initiate PayPal Subscription...<br><br>${error}`,
            );
        }
    },
    onApprove(data) {
        /*
        No need to activate manually since SUBSCRIBE_NOW is being used.
        Learn how to handle other user actions from our docs:
        https://developer.paypal.com/docs/api/subscriptions/v1/#subscriptions_create
        */
        if (data.orderID) {
            resultMessage(
                `You have successfully subscribed to the plan. Your subscription id is: ${data.subscriptionID}`,
            );
        } else {
            resultMessage(
                `Failed to activate the subscription: ${data.subscriptionID}`,
            );
        }
    },
}).render("#paypal-button-container-subscribe"); // Renders the PayPal button
