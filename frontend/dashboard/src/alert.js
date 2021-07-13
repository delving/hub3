let delayedFade;

export function alert(message, isError) {
  const alertContainer = document.getElementById('alert');
  alertContainer.classList.remove('fade-out');
  alertContainer.textContent = message;
  const bgColor = isError ? 'red' : 'blue';
  alertContainer.style.backgroundColor = bgColor;
  clearTimeout(delayedFade);
  delayedFade = setTimeout(() => {
    alertContainer.classList.add('fade-out');
  }, 3000);
}