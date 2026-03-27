export const extractFilename = (contentDisposition, defaultName = "download") => {
    if (!contentDisposition) return defaultName;

    // Tenta formato RFC 5987 (filename*=UTF-8'') - case insensitive
    let match = contentDisposition.match(/filename\*=UTF-8''([^;]+)/i);
    if (match && match[1]) {
        try {
            return decodeURIComponent(match[1]);
        } catch (e) {
            console.error('Failed to decode filename:', e);
        }
    }

    // Tenta formato Quoted (filename="name.ext")
    match = contentDisposition.match(/filename="([^"]+)"/i);
    if (match && match[1]) return match[1];

    // Tenta formato sem aspas (filename=name.ext)
    match = contentDisposition.match(/filename=([^;\r\n"']+)/i);
    if (match && match[1]) return match[1].trim();

    return defaultName;
};
