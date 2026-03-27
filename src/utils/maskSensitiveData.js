export const maskCpf = (cpf, hasPermission) => {
    if (!cpf) return 'N/A';

    // Remove formatting if any
    const cleanCpf = cpf.replace(/[^\d]/g, '');

    // Layout and Mask logic
    if (cleanCpf.length === 11) {
        if (hasPermission) {
            return `${cleanCpf.substring(0, 3)}.${cleanCpf.substring(3, 6)}.${cleanCpf.substring(6, 9)}-${cleanCpf.substring(9, 11)}`;
        }
        return `***.${cleanCpf.substring(3, 6)}.${cleanCpf.substring(6, 9)}-**`;
    }

    return hasPermission ? cpf : '***.***.***-**';
};
