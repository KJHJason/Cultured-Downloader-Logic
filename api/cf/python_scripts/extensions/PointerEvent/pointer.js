// from https://github.com/TheFalloutOf76/CDP-bug-MouseEvent-.screenX-.screenY-patcher/tree/main
Object.defineProperty(
    MouseEvent.prototype, 
    "screenX", 
    { value: 1234 },
);
Object.defineProperty(
    MouseEvent.prototype,
    "screenY", 
    { value: 567 },
);
