function isFormValid(form) {
  return form.checkValidity();
}

function removePayPalElement() {
  const element = document.getElementById('paypal-button-container');
  element.innerHTML = '';
  element.style.display = 'none';
}

function hideDialog() {
  document.getElementById('overlay').style.display = 'none';
  document.getElementById('dialog').style.display = 'none';
  document.getElementById('openDialogButton').style.display = 'block';
}

function renderPaypalElement() {
  const element = document.getElementById('paypal-button-container');
  element.innerHTML = '';
  element.style.display = 'block';

  paypal.Buttons({
    style: {color: 'blue', shape: 'pill', label: 'pay', height: 40},

    // Call your server to set up the transaction
    createOrder: function(data, actions) {
      return fetch('/demo/checkout/api/paypal/order/create/', {
        method: 'post'
      }).then(function(res) {
        return res.json();
      }).then(function(orderData) {
        return orderData.id;
      });
    },

    // Call your server to finalize the transaction
    onApprove: function(data, actions) {
      return fetch('/demo/checkout/api/paypal/order/' + data.orderID + '/capture/', {
        method: 'post'
      }).then(function(res) {
        return res.json();
      }).then(function(orderData) {
        var errorDetail = Array.isArray(orderData.details) && orderData.details[0];

        //   (1) Recoverable INSTRUMENT_DECLINED -> call actions.restart()
        if (errorDetail && errorDetail.issue === 'INSTRUMENT_DECLINED') {
          return actions.restart();
        }

        //   (2) Other non-recoverable errors -> Show a failure message
        if (errorDetail) {
          var msg = 'Sorry, your transaction could not be processed.';
          if (errorDetail.description) msg += '\n\n' + errorDetail.description;
          if (orderData.debug_id) msg += ' (' + orderData.debug_id + ')';
          return alert(msg); // TODO show a prettier message
        }

        //   (3) Successful transaction -> Show confirmation or thank you
        // Grab transaction.status and transaction.id, call up php and save it in db.
        // var transaction = orderData.purchase_units[0].payments.captures[0];
        // alert('Transaction '+ transaction.status + ': ' + transaction.id + '\n\nSee console for all available details');

        // Replace the above to show a success message within this page, e.g.
        element.innerHTML = '';
        element.innerHTML = '<h3>Thank you for your payment!</h3>';
        document.getElementById('mainForm').submit();
      });
    }
  }).render('#paypal-button-container');
}

document.getElementById('openDialogButton').addEventListener('click', () => {
  document.getElementById('overlay').style.display = 'block';
  document.getElementById('dialog').style.display = 'block';
  document.getElementById('openDialogButton').style.display = 'none';
});

document.getElementById('submitDialogButton').addEventListener('click', () => {
  const form = document.getElementById('mainForm');

  ['name', 'email', 'phone'].forEach(id => {
    const input = document.createElement('input');
    input.type = 'hidden';
    input.name = id;
    input.value = document.getElementById(id).value;
    form.appendChild(input);
  });

  if (isFormValid(form)) {
    document.getElementById('submitDialogButton').style.display = 'none';
    document.getElementById('error-message').style.display = 'none';
    renderPaypalElement();
  } else {
    document.getElementById('submitDialogButton').style.display = 'inline';
    document.getElementById('error-message').style.display = 'block';
    removePayPalElement();
  }

  // const orderIDInput = document.createElement('input');
  // orderIDInput.type = 'hidden';
  // orderIDInput.name = 'paypalOrderID';
  // orderIDInput.value = paypalOrderID;
  // form.appendChild(orderIDInput);
  // form.submit();
});

document.getElementById('cancelDialogButton').addEventListener('click', () => {
  hideDialog();
});

document.addEventListener('click', (event) => {
  const dialog = document.getElementById('dialog');
  const openDialogButton = document.getElementById('openDialogButton');
  // If the click is outside the dialog and not on the openDialogButton, hide the dialog
  if (!dialog.contains(event.target) && !openDialogButton.contains(event.target)) {
    hideDialog();
  }
});
