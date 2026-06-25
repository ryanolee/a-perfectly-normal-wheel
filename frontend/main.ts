import "./styles/style.css";
import 'htmx.org';
import 'htmx.org/dist/ext/sse.js';
import 'htmx.org/dist/ext/remove-me.js';

async function spinWheel(candidates: string[], winner: string) {
    const [{ spinTheWheel }, { realisticBlast }] = await Promise.all([
        import('./winner/wheel'),
        import('./winner/confetti')
    ]);
    
    await spinTheWheel(candidates, winner);
    realisticBlast();
    setTimeout(() => location.reload(), 3000);
}

(window as any).spinWheel = spinWheel;
