document.addEventListener('DOMContentLoaded', () => {
// Smart Navbar Scroll Effect
const navbar = document.getElementById('navbar');
    let lastScrollY = window.scrollY;
    
    window.addEventListener('scroll', () => {
        // Scrolled state for background
        if (window.scrollY > 20) {
            navbar.classList.add('scrolled');
        } else {
            navbar.classList.remove('scrolled');
        }


    });
});
