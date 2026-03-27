const fs = require('fs');
const files = [
    '/Pixelcraft-studio-website/src/components/admin/roles/ExportImportModal.jsx',
    '/Pixelcraft-studio-website/src/components/admin/roles/NotificationToast.jsx',
    '/Pixelcraft-studio-website/src/components/dashboard/DepositModal.jsx',
    '/Pixelcraft-studio-website/src/pages/Support.jsx',
    '/Pixelcraft-studio-website/src/pages/admin/AdminCatalog.jsx'
];

for (const file of files) {
    let content = fs.readFileSync(file, 'utf8');

    // The bad replacement omitted the angle bracket `>` and appended `$5`.
    // It looks like: aria-label="Fechar"\n <Icon />\n </button>$5
    // We need to restore the `>` before the <Icon /> and remove `$5`.

    // Find aria-label="X" followed by an icon and closing button with $5
    const brokenRegex = /aria-label="([^"]+)"(\s*<[A-Z][A-Za-z0-9]*\b[\s\S]*?\/>\s*<\/(?:motion\.)?button>)\$5/g;

    content = content.replace(brokenRegex, 'aria-label="$1">$2');

    fs.writeFileSync(file, content, 'utf8');
    console.log('Fixed', file);
}
