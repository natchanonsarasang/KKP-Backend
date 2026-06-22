const fs = require('fs');
const path = require('path');
const YAML = require('yaml');

try {
  const yamlPath = path.join(__dirname, '../openapi.yaml');
  const yamlText = fs.readFileSync(yamlPath, 'utf8');
  const parsedSpec = YAML.parse(yamlText);
  const jsonSpec = JSON.stringify(parsedSpec, null, 2);

  const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Callecto API - Interactive Reference Guide</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
  <link rel="shortcut icon" href="https://fastly.jsdelivr.net/npm/swagger-ui-dist@5/favicon-32x32.png" />
  <style>
    html {
      box-sizing: border-box;
      overflow: -moz-scrollbars-vertical;
      overflow-y: scroll;
    }
    *, *:before, *:after {
      box-sizing: inherit;
    }
    body {
      margin: 0;
      background: #fafafa;
    }
    .swagger-ui .topbar {
      background-color: #1b1c1d;
    }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js" charset="UTF-8"></script>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js" charset="UTF-8"></script>
  <script>
    window.onload = () => {
      window.ui = SwaggerUIBundle({
        spec: ${jsonSpec},
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "BaseLayout"
      });
    };
  </script>
</body>
</html>
`;

  const docsDir = path.join(__dirname, '../docs');
  if (!fs.existsSync(docsDir)) {
    fs.mkdirSync(docsDir);
  }

  const htmlPath = path.join(docsDir, 'interactive_docs.html');
  fs.writeFileSync(htmlPath, htmlTemplate, 'utf8');
  console.log('docs/interactive_docs.html generated successfully.');
} catch (err) {
  console.error('Error generating interactive docs:', err);
  process.exit(1);
}
