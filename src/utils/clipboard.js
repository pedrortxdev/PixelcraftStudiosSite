/**
 * Fallback para o clipboard em conexões HTTP inseguras ou navegadores antigos.
 * (BUG-124) T-068
 */
export const copyToClipboard = async (text) => {
    if (navigator.clipboard && navigator.clipboard.writeText) {
        try {
            await navigator.clipboard.writeText(text);
            return true;
        } catch (err) {
            console.error('Failed to write to clipboard via API', err);
        }
    }

    // Fallback: Try using the deprecated document.execCommand
    try {
        const textArea = document.createElement("textarea");
        textArea.value = text;

        // Avoid scrolling to bottom
        textArea.style.top = "0";
        textArea.style.left = "0";
        textArea.style.position = "fixed";

        document.body.appendChild(textArea);
        textArea.focus();
        textArea.select();

        const successful = document.execCommand('copy');
        document.body.removeChild(textArea);
        return successful;
    } catch (err) {
        console.error('Fallback clipboard failed', err);
        return false;
    }
};
