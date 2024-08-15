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
        document.getElementById('paypalOneTimeButton').classList.add('active');
    } else if (selectedButtonId === 'showSubButton') {
        document.getElementById('paypal-button-container').classList.add('active');
        document.getElementById('paypalSubButton').classList.add('active');
    }
}

function isFormValid(form) {
  return form.checkValidity();
}

paypal_onetime.Buttons({
    style: { shape: 'pill', color: 'black', layout: 'vertical', label: 'pay' },
    createOrder: function(data, actions) {
        return actions.order.create({
            intent: 'CAPTURE',
            purchase_units: [{
                amount: {
                    currency_code: 'USD',
                    value: '20.00'
                }
            }]
        });
    },
    onApprove: function(data, actions) {
        return actions.order.capture().then(function(details) {
            alert('Transaction completed by ' + details.payer.name.given_name);
        });
    }
}).render("#paypalOneTimeButton");

paypal_subscribe.Buttons({
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
