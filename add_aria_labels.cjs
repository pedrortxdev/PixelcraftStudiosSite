const fs = require('fs');
const path = require('path');

const srcDir = '/Pixelcraft-studio-website/src';

const iconLabels = {
    'X': 'Fechar',
    'Trash2': 'Excluir',
    'Edit': 'Editar',
    'Edit2': 'Editar',
    'Menu': 'Menu',
    'Search': 'Buscar'
};

function processFile(filePath) {
    let content = fs.readFileSync(filePath, 'utf8');
    let original = content;

    for (const [icon, label] of Object.entries(iconLabels)) {
        // Regex for <button ...> <Icon /> </button>
        const regex = new RegExp(`(<(motion\\.)?button(?!.*aria-label)[^>]*?)(>)(\\s*<${icon}\\b[\\s\\S]*?/>\\s*</(?:motion\\.)?button>)`, 'g');
        content = content.replace(regex, `$1 aria-label="${label}"$4$5`);

        // Regex for <div onClick ...> <Icon /> </div>
        const divRegex = new RegExp(`(<div(?!.*aria-label)[^>]*?onClick={[^>]*?)(>)(\\s*<${icon}\\b[\\s\\S]*?/>\\s*</div>)`, 'g');
        content = content.replace(divRegex, `$1 aria-label="${label}" role="button" tabIndex={0}$3$4`);
    }

    // Specific case for standalone icons with onClick acts as a button
    const xRegex = new RegExp(`(<X(?!.*aria-label)[^>]*?onClick={[^>]*?)(/?>)`, 'g');
    content = content.replace(xRegex, `$1 aria-label="Fechar" role="button" tabIndex={0}$3`);


    if (content !== original) {
        fs.writeFileSync(filePath, content, 'utf8');
        console.log(`Updated aria-labels in: ${filePath.replace(srcDir, '')}`);
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
console.log('Aria label script executed.');
