function initializeEditor(conex_data) {
    directory = conex_data?.directory || "temp";

    editor = new EditorJS({
        // readOnly: false,
        holder: 'editorjs',

        inlineToolbar: ['marker', 'bold', 'italic'],
        inlineToolbar: true,
        tools: {
            /**
             * Each Tool is a Plugin. Pass them via 'class' option with necessary settings {@link docs/tools.md}
             */
            header: {
                class: Header,
                config: {
                    placeholder: 'Inserta un título',
                    levels: [1, 2, 3],
                    defaultLevel: 1,
                    shortcut: 'CMD+SHIFT+H'
                }
            },

            image: {
                class: ImageTool,
                config: {
                    endpoints: {
                        byFile: `https://api.conex.one/api/upload`,
                    },
                    field: 'file',
                    types: 'image/*',
                    additionalRequestData: {
                        directory: directory,
                    },
                },
            },

            list: {
                class: List,
                inlineToolbar: true,
                shortcut: 'CMD+SHIFT+L'
            },

            quote: {
                class: Quote,
                inlineToolbar: true,
                config: {
                    quotePlaceholder: 'Insertar una cita',
                    captionPlaceholder: 'Autor de la cita',
                },
                shortcut: 'CMD+SHIFT+O'
            },

            table: {
                class: Table,
                inlineToolbar: true,
                shortcut: 'CMD+ALT+T'
            },
        },

        data: conex_data?.editor_data || {
            blocks: [
                {
                    type: "header",
                    data: {
                        text: "Acerca de [Empresa]",
                        level: 1
                    }
                },
                {
                    type : 'paragraph',
                    data : {
                        text : 'En [Nombre de Tu Empresa], nos dedicamos a ofrecer [tu servicio/producto] de la más alta calidad con un servicio al cliente excepcional. Nuestro equipo de expertos se asegura de que cada aspecto de tu experiencia sea manejado con profesionalismo y cuidado.'
                    }
                },
                {
                    type : 'list',
                    data : {
                        items : [
                            'Resolvemos una necesidad clave de mercado',
                            'Inversión en crecimiento con presupuesto sostenible.',
                            'Enfoque en satisfacción del cliente',
                        ],
                        style: 'unordered'
                    }
                },
                {
                    type: 'table',
                    data: {
                        content: [
                            ['Servicios', 'Descripción', 'Costo'],
                            ['Impresión', 'Breve descripción', '1000'],
                            ['laminado', 'Breve descripción', '2000'],
                        ]
                    }
                },
            ]
        },
        i18n: {
            messages: {
                ui: {
                    "blockTunes": {
                        "toggler": {
                            "Click to tune": "Modificar",
                            "or drag to move": "or drag to move"
                        },
                    },
                    "inlineToolbar": {
                        "converter": {
                            "Convert to": "Convertir a"
                        }
                    },
                    "toolbar": {
                        "toolbox": {
                            "Add": "Insertar"
                        }
                    }
                },

                /**
                 * Section for translation Tool Names: both block and inline tools
                 */
                toolNames: {
                    "Text": "Texto",
                    "Heading": "Título",
                    "List": "Lista",
                    "Warning": "Advertencia",
                    "Quote": "Cita",
                    "Table": "Tabla",
                    "Link": "Link",
                    "Image": "Imagen",
                    "Bold": "Negrita",
                    "Italic": "Itálicas",
                    "InlineCode": "InlineCode",
                },

                /**
                 * TRANSLATIONS
                 * Each subsection is the i18n dictionary that will be passed to the corresponded plugin
                 * The name of a plugin should be equal the name you specify in the 'tool' section for that plugin
                 */
                tools: {
                    "warning": { // <-- 'Warning' tool will accept this dictionary section
                        "Title": "Título",
                        "Message": "Mensaje",
                    },

                    "link": {
                        "Add a link": "Agregar link"
                    },
                    "stub": {
                        'The block can not be displayed correctly.': 'No se puede visualizar este bloque'
                    }
                },

                blockTunes: {
                    /**
                     * Each subsection is the i18n dictionary that will be passed to the corresponded Block Tune plugin
                     * The name of a plugin should be equal the name you specify in the 'tunes' section for that plugin
                     *
                     * Also, there are few internal block tunes: "delete", "moveUp" and "moveDown"
                     */
                    "delete": {
                        "Delete": "Quitar bloque"
                    },
                    "moveUp": {
                        "Move up": "Mover arriba"
                    },
                    "moveDown": {
                        "Move down": "Mover abajo"
                    }
                },
            }
        },
        onChange: function(api, event) {
            saveEditorData();
        }
    });
}
