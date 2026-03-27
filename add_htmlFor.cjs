const fs = require('fs');
const path = require('path');

const srcDir = '/Pixelcraft-studio-website/src';

function processFile(filePath) {
    let content = fs.readFileSync(filePath, 'utf8');
    let original = content;

    // Finds <label ...> ... </label> \s* <input/select/textarea ... name="foo" ... >
    // Adds htmlFor="foo" to label if missing, and id="foo" to input if missing.
    const regex = /<label((?!htmlFor)[^>]*)>([\s\S]*?)<\/label>\s*<(input|select|textarea)([^>]*?)name=["']([^"']+)["']([^>]*?)>/g;

    content = content.replace(regex, (match, labelAttrs, labelContent, tagName, beforeName, nameValue, afterName) => {
        let newLabelAttrs = ` htmlFor="${nameValue}"` + labelAttrs;

        let newBefore = beforeName;
        // Check if id is already present anywhere in the input tag
        const fullTag = `<${tagName}${beforeName}name="${nameValue}"${afterName}>`;
        if (!fullTag.includes('id=')) {
            newBefore = ` id="${nameValue}"` + beforeName;
        }

        return `<label${newLabelAttrs}>${labelContent}</label>\n<(tagName==="textarea" || tagName==="select" || tagName==="input" ? tagName : "") ${newBefore}name="${nameValue}"${afterName}>`.replace(/<\(.*?\) /g, `<${tagName} `);
    });

    // Also run for inputs that have value={...} but maybe don't have name (less common, but possible)

    if (content !== original) {
        fs.writeFileSync(filePath, content, 'utf8');
        console.log(`Updated htmlFor in: ${filePath.replace(srcDir, '')}`);
    }
}

function walkDir(dir) {
    const files = fs.readdirSync(dir);
    for (const file of files) {
        if (file === 'node_modules') continue;
        const fullPath = path.join(dir, file);
        if (fs.statSync(fullPath).isDirectory()) {
            walkDir(fullPath);
        } else if (fullPath.endsWith('.jsx') || fullPath.endsWith('.js')) {
            processFile(fullPath);
        }
    }
}

walkDir(srcDir);
console.log('htmlFor script executed.');
