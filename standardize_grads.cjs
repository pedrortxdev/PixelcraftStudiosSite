const fs = require('fs');
const path = require('path');

const directoryPath = path.join(__dirname, 'src');

function findAndReplaceGrads(dir) {
    const files = fs.readdirSync(dir);

    for (const file of files) {
        const fullPath = path.join(dir, file);
        if (fs.statSync(fullPath).isDirectory()) {
            findAndReplaceGrads(fullPath);
        } else if (fullPath.endsWith('.jsx') || fullPath.endsWith('.js') || fullPath.endsWith('.tsx') || fullPath.endsWith('.ts')) {
            let content = fs.readFileSync(fullPath, 'utf8');
            let modified = false;

            // Primary gradients
            const primaryRegex = /'linear-gradient\(135deg,\s*(?:#583AFF\s+0%,\s*#1AD2FF\s+100%|#583AFF,\s*#1AD2FF)\)'/g;
            if (primaryRegex.test(content)) {
                content = content.replace(primaryRegex, "'var(--gradient-primary)'");
                modified = true;
            }

            // CTA gradients
            const ctaRegex = /'linear-gradient\(135deg,\s*(?:#E01A4F\s+0%,\s*#FF6B35\s+100%|#E01A4F,\s*#FF6B35)\)'/g;
            if (ctaRegex.test(content)) {
                content = content.replace(ctaRegex, "'var(--gradient-cta)'");
                modified = true;
            }

            if (modified) {
                fs.writeFileSync(fullPath, content, 'utf8');
                console.log(`Updated gradients in ${fullPath}`);
            }
        }
    }
}

findAndReplaceGrads(directoryPath);
console.log('Gradient standardization complete.');
