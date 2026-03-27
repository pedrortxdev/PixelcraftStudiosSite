const fs = require('fs');
const path = require('path');

const srcDir = '/Pixelcraft-studio-website/src';

function processFile(filePath) {
    let content = fs.readFileSync(filePath, 'utf8');
    let original = content;

    // Standardize paddings
    content = content.replace(/padding:\s*['"]0\.5rem\s+1rem['"]/g, "padding: 'var(--btn-padding-sm)'");
    content = content.replace(/padding:\s*['"]0\.75rem\s+1\.5rem['"]/g, "padding: 'var(--btn-padding-md)'");
    content = content.replace(/padding:\s*['"]1rem\s+2rem['"]/g, "padding: 'var(--btn-padding-lg)'");
    content = content.replace(/padding:\s*['"]1\.5rem\s+3rem['"]/g, "padding: 'var(--btn-padding-xl)'");
    // Other common sizes in codebase
    content = content.replace(/padding:\s*['"]12px\s+24px['"]/g, "padding: 'var(--btn-padding-md)'");
    content = content.replace(/padding:\s*['"]1rem['"]/g, "padding: 'var(--btn-padding-md)'");

    // Standardize title font sizes
    content = content.replace(/fontSize:\s*['"]4rem['"]/g, "fontSize: 'var(--title-hero)'");
    content = content.replace(/fontSize:\s*['"]3rem['"]/g, "fontSize: 'var(--title-h1)'");
    content = content.replace(/fontSize:\s*['"]2\.5rem['"]/g, "fontSize: 'var(--title-h2)'");
    content = content.replace(/fontSize:\s*['"]2rem['"]/g, "fontSize: 'var(--title-h3)'");
    content = content.replace(/fontSize:\s*['"]1\.5rem['"]/g, "fontSize: 'var(--title-h4)'");

    if (content !== original) {
        fs.writeFileSync(filePath, content, 'utf8');
        console.log(`Updated: ${filePath.replace(srcDir, '')}`);
    }
}

function walkDir(dir) {
    const files = fs.readdirSync(dir);
    for (const file of files) {
        const fullPath = path.join(dir, file);
        if (fs.statSync(fullPath).isDirectory()) {
            walkDir(fullPath);
        } else if (fullPath.endsWith('.jsx') || fullPath.endsWith('.js')) {
            processFile(fullPath);
        }
    }
}

walkDir(srcDir);
console.log('UI Standardization script executed.');
