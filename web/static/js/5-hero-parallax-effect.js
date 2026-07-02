document.addEventListener('DOMContentLoaded', () => {
// 5 Hero Parallax Effect
const heroImg = document.querySelector('.hero-image img');
    if (heroImg) {
        window.addEventListener('scroll', () => {
            const scroll = window.scrollY;
            if (scroll < 800) {
                heroImg.style.transform = `translateY(${scroll * 0.12}px) scale(1.02)`;
            }
        });
    }
});
