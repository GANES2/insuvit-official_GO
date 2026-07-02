document.addEventListener('DOMContentLoaded', () => {
// Scroll Reveal Animation using Intersection Observer
const fadeElements = document.querySelectorAll('.fade-in, .fade-in-left, .fade-in-right');
    
    const revealOptions = {
        threshold: 0.15,
        rootMargin: "0px 0px -50px 0px"
    };

    const revealOnScroll = new IntersectionObserver(function(entries, observer) {
        entries.forEach(entry => {
            if (!entry.isIntersecting) return;
            
            entry.target.classList.add('visible');
            observer.unobserve(entry.target);
        });
    }, revealOptions);

    fadeElements.forEach(el => {
        revealOnScroll.observe(el);
    });
});
