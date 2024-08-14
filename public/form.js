// TODO
//
// 1. See if I can stop using hosted buttons and use
//    https://developer.paypal.com/integration-builder/ instead
//    to use components=buttons only. Or, Try to load both,
//    components=buttons,hosted-buttons etc.
//    See Buttons SDK Reference
//    https://developer.paypal.com/sdk/js/reference/
//
// 2. Try to disable asking for shipping info (although could be
//    useful to mark as sent).
//
// 3. Read about IPN and Webhooks to automate registering process.

const PayPalSdkOneTime = "";
const PayPalSdkSub = "";
const clientId = "";
const OneTimePID = "";
const PlanID = "";

loadOneTimeButton().then(() => {loadSubButton()})

function loadOneTimeButton() {
    return new Promise((resolve, reject) => {
        var script = document.createElement('script');
        script.src = PayPalSdkOneTime;
        script.onload = function() {
            paypal.HostedButtons({
                hostedButtonId: OneTimePID,
            }).render("#paypalOneTimeButton");
        };
        document.head.appendChild(script);
        resolve();
    });
}

function loadSubButton() {
    var script = document.createElement('script');
    script.src = PayPalSdkSub;
    script.onload = function() {
        paypal.Buttons({
            style: { shape: 'pill', color: 'black', layout: 'vertical', label: 'subscribe' },
            createSubscription: function(data, actions) {
                return actions.subscription.create({
                    plan_id: PlanID
                });
            },
            onApprove: function(data, actions) {
                alert(data.subscriptionID); // You can add optional success message for the subscriber here
            }
        }).render('#paypalSubButton');
    };
    document.head.appendChild(script);
}

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
        document.getElementById('paypalOneTimeButton').classList.add('active');
    } else if (selectedButtonId === 'showSubButton') {
        document.getElementById('paypal-button-container').classList.add('active');
        document.getElementById('paypalSubButton').classList.add('active');
    }
}

document.getElementById('showOneTimeButton').addEventListener('click', function() {
    togglePaymentMethod('showOneTimeButton');
});

document.getElementById('showSubButton').addEventListener('click', function() {
    togglePaymentMethod('showSubButton');
});

document.getElementById('openDialogButton').addEventListener('click', () => {
    showDialog();
});

document.getElementById('cancelDialogButton').addEventListener('click', () => {
    hideDialog();
});
