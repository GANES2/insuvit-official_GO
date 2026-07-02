document.addEventListener('DOMContentLoaded', () => {
    // FAQ Accordion (with ARIA for accessibility)
    const faqQuestions = document.querySelectorAll('.faq-question');

    faqQuestions.forEach((question, index) => {
        const answer = question.nextElementSibling;
        const answerId = 'faq-answer-' + index;
        answer.id = answerId;
        answer.setAttribute('role', 'region');
        question.setAttribute('aria-expanded', 'false');
        question.setAttribute('aria-controls', answerId);

        question.addEventListener('click', () => {
            const faqItem = question.parentElement;
            const isActive = faqItem.classList.contains('active');

            document.querySelectorAll('.faq-item').forEach(item => {
                item.classList.remove('active');
                if (item.querySelector('.faq-answer')) {
                    item.querySelector('.faq-answer').style.maxHeight = null;
                }
                if (item.querySelector('.faq-question')) {
                    item.querySelector('.faq-question').setAttribute('aria-expanded', 'false');
                }
            });

            if (!isActive) {
                faqItem.classList.add('active');
                answer.style.maxHeight = answer.scrollHeight + "px";
                question.setAttribute('aria-expanded', 'true');
            }
        });
    });
});
