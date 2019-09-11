
export function formatWebSocketUri(host : string) : string {
    return host.replace(/^http/, 'ws')
}