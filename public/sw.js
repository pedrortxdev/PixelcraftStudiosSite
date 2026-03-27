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

    // ESTRATÉGIA: Network-First para navegação (index.html)
    // Isso garante que o usuário sempre pegue a versão mais recente do HTML
    if (event.request.mode === 'navigate') {
        event.respondWith(
            fetch(event.request)
                .then((response) => {
                    // Opcional: Atualizar o cache com a nova versão
                    const copy = response.clone();
                    caches.open(CACHE_NAME).then((cache) => cache.put(event.request, copy));
                    return response;
                })
                .catch(() => {
                    // Se estiver offline, tenta o cache
                    return caches.match(event.request);
                })
        );
        return;
    }

    // ESTRATÉGIA: Cache-First para outros recursos estáticos (JS, CSS, Imagens)
    event.respondWith(
        caches.match(event.request).then((response) => {
            if (response) return response;

            return fetch(event.request).then((networkResponse) => {
                // Não cachear respostas de erro ou de outros domínios aqui (já filtrado acima)
                if (!networkResponse || networkResponse.status !== 200) return networkResponse;
                
                const responseToCache = networkResponse.clone();
                caches.open(CACHE_NAME).then((cache) => {
                    cache.put(event.request, responseToCache);
                });

                return networkResponse;
            }).catch(() => null);
        })
    );
});
