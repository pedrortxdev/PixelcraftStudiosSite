const CACHE_NAME = 'pixelcraft-pwa-v1';
const ASSETS_TO_CACHE = [
    '/',
    '/index.html',
    '/manifest.json',
    '/vite.svg'
];

self.addEventListener('install', (event) => {
    event.waitUntil(
        caches.open(CACHE_NAME).then((cache) => cache.addAll(ASSETS_TO_CACHE))
    );
    self.skipWaiting();
});

self.addEventListener('activate', (event) => {
    event.waitUntil(
        caches.keys().then((cacheNames) => {
            return Promise.all(
                cacheNames.map((cache) => {
                    if (cache !== CACHE_NAME) return caches.delete(cache);
                })
            );
        })
    );
    self.clients.claim();
});

self.addEventListener('fetch', (event) => {
    // Apenas GET e mesma origem (evita erros com Cloudflare/Scripts externos)
    if (event.request.method !== 'GET') return;
    
    const url = new URL(event.request.url);
    
    // IGNORAR COMPLETAMENTE: API, Admin e domínios externos (Cloudflare, etc)
    if (url.pathname.includes('/api/') || 
        url.pathname.includes('/admin/') || 
        url.hostname.includes('api.pixelcraft-studio.store') ||
        url.hostname !== self.location.hostname) {
        return;
    }

    event.respondWith(
        caches.match(event.request).then((response) => {
            if (response) return response;

            return fetch(event.request).catch(() => {
                // Se falhar a navegação (offline), mostra o index.html
                if (event.request.mode === 'navigate') {
                    return caches.match('/index.html');
                }
                // Para outros arquivos, apenas deixa falhar silenciosamente
                return null;
            });
        })
    );
});
