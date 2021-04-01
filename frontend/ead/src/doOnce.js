export function doOnce(target, eventType, handler) {
  const onceHandler = () => {
    target.removeEventListener(eventType, onceHandler);
    handler();
  };
  target.addEventListener(eventType, onceHandler, {
    capture: true
  });
};