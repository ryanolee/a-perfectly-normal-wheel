import * as PIXI from 'pixi.js';
import { gsap } from 'gsap';
import { PixiPlugin } from 'gsap/PixiPlugin';


gsap.registerPlugin(PixiPlugin);
PixiPlugin.registerPIXI(PIXI);

const rotationSpeed = 0.25;
const rotationDeceleration = 0.001;
const rotationDecelerationDecay = 0.000002;
const rotationDecelerationMin = 0.0001;
const randomSpeedOffset = 0.02 * Math.random();


enum WheelState {
    Spinning,
    ShiftingWinners
}

export const spinTheWheel = (names: string[], winner: string,) => {
    return new Promise<void>((resolve) => {
        setupSpin(names, winner, resolve);
    })
}

const setupSpin = async (names: string[], winner: string,  onComplete = () => {}) => {
    let wheelState: WheelState = WheelState.Spinning;
    const container = document.querySelector('#wheelContainer') as HTMLElement;
    const app = new PIXI.Application()

    await app.init({ backgroundAlpha: 0, resizeTo: container, antialias: true })
    container.appendChild(app.canvas);

   

    const wheelRadius = Math.min(app.screen.width, app.screen.height) / 3;

    const arcs = getArcs(names.length, wheelRadius, app.screen.width / 2, app.screen.height / 2);
    arcs.forEach(arc => app.stage.addChild(arc));

    const nameElements = getNamesInCircle( names, wheelRadius * 0.7 , app.screen.width / 2, app.screen.height / 2);
    nameElements.forEach(name => app.stage.addChild(name));

    const gopher = await getGopher(app.screen.width / 2, app.screen.height / 2, wheelRadius);
    
    app.stage.addChild(gopher);

    let gopherRotationSpeed = rotationSpeed + randomSpeedOffset;
    let gopherRotationDeceleration = rotationDeceleration;
    
    const spinWheel = () => {
        if(wheelState !== WheelState.Spinning) {
            return;
        }

        gopher.rotation += gopherRotationSpeed;
        gopher.rotation %= 2 * Math.PI;

        gopherRotationSpeed -= gopherRotationDeceleration;
        gopherRotationDeceleration = Math.max(gopherRotationDeceleration - rotationDecelerationDecay, rotationDecelerationMin);
        if (gopherRotationSpeed < 0) {
            app.ticker.remove(spinWheel);
            shiftWinners();
        }
    }

    const shiftWinners = () => {
        // Get closet rotationally to the gopher
        const winnerElement = nameElements.reduce((closest, current) => {
            const targetRotation = (gopher.rotation + Math.PI) % (2 * Math.PI);

            const closestDiff = angularDiff(closest.rotation, targetRotation);
            const currentDiff = angularDiff(current.rotation, targetRotation);
            return currentDiff < closestDiff ? current : closest;
            
        });
        const realWinnerElement = nameElements.find(name => name.text === winner);
        if (realWinnerElement?.text === winnerElement.text) {
            onComplete();
            return;
        }

        swapText(winnerElement, realWinnerElement!, onComplete );
    }

    app.ticker.add(spinWheel);
}

const swapText = (a: PIXI.Text, b: PIXI.Text, onComplete = () => {}) => {
    gsap.to(a, {
        pixi: { x: b.x, y: b.y, rotation: radToDeg(b.rotation) },
        duration: 5,
    })
    gsap.to(b, {
        pixi: { x: a.x, y: a.y, rotation: radToDeg(a.rotation) },
        duration: 5,
        onComplete: onComplete
    })
}


const radToDeg = (rad: number) => rad * (180 / Math.PI);
const angularDiff = (a: number, b: number) => {
    const diff = Math.abs(a - b) % (2 * Math.PI);
    return Math.min(diff, 2 * Math.PI - diff);
};

const getArcs = (segments: number, radius: number, centerX: number, centerY: number) => {
    const angleStep = (2 * Math.PI) / segments;
    
    return Array.from({ length: segments }, (_, index) => {
        const angle = (index * angleStep) + angleStep / 2; // Center the arc on the segment
        const arc = new PIXI.Graphics();
         arc.moveTo(centerX, centerY)
           .arc(centerX, centerY, radius, angle, angle + angleStep)
           .closePath()
           //.fill({ color: 0x000000, alpha: 0.5 })
           .stroke({ color: 0xffffff, width: 2 });

        return arc;
    })
}
    


const getNamesInCircle = (names: string[], radius: number, centerX: number, centerY: number) => {
    const angleStep = (2 * Math.PI) / names.length;

    return names.map((name, index) => {
        const angle = index * angleStep;


        const text = new PIXI.Text({
            text: name,
            style: {
                fontFamily: 'Arial',
                fontSize: 24,
                fill: 0xffffff,
            }
        });
        text.anchor.set(0.5);
        text.x = centerX + radius * Math.cos(angle);
        text.y = centerY + radius * Math.sin(angle);

        text.rotation = angle ;
        return text
    })
}

const getGopher = async (x: number, y: number, lineOfSightLength: number) => {
    const gopherContainer = new PIXI.Container();
    gopherContainer.x = x
    gopherContainer.y = y

    const gopherAsset = await PIXI.Assets.load('/img/gopher.png')
    const gopher = new PIXI.Sprite(gopherAsset);

    gopher.anchor.set(0.45, 0.5);
    gopher.scale.set(0.1);

    const visionLine = new PIXI.Graphics();
    const dashLength = 15;
    const gapLength = 10;
    for (let x = 0; x < lineOfSightLength; x += dashLength + gapLength) {
        visionLine.moveTo(x, 0).lineTo(Math.min(x + dashLength, lineOfSightLength), 0);
    }
    visionLine.stroke({ color: 0xffffff, width: 2, alpha: 0.8 });
    visionLine.rotation = Math.PI
    gopherContainer.addChild(visionLine);
    gopherContainer.addChild(gopher);

    return gopherContainer;
}

        
