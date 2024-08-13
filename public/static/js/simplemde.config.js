var simplemde = new SimpleMDE({
  element: document.getElementById("editor"),
  toolbar: ["preview", "|", "heading", "bold", "italic", "unordered-list", "ordered-list", "|", "link", "image", "table"],
  spellChecker: false,
});
simplemde.value("# Bienvenido a [Empresa]\n\nUbicados en [Dirección de la Empresa]\n\n✉️ contacto@empresa.com\n☎️ +506 8888 8888\n\n# Servicios\n\n| Servicio | Descripción | Monto     |\n| -------- | ----------- | --------- |\n| Uno      | Una         |  1.000,00 |\n| Dos      | Breve       | 10.000,00 |\n| Tres     | Explicación |  7.500,00 |\n\n![Imagen ejemplo](https://0x0.st/XWHZ.jpg)\n\n# Acerca de Nosotros\n\nEn [Empresa], nos especializamos en [breve descripción de tus servicios/productos]. Nuestro equipo está dedicado a ofrecer [propuesta de valor o punto de venta único].\n\n# Síguenos\n\n[Facebook](https://facebook.com) | [Instagram](https://instagram.com)\n");
document.getElementById('openDialogButton').addEventListener('click', () => {
  document.getElementById('dialog').style.display = 'block';
});
