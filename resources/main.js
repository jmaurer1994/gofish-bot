showSlides("weather-icon", 1);
showSlides("weather-info", 1);
function showSlides(c, si) {
    let i;
    let slides = document.getElementsByClassName(c);
    let slideIndex = si;
    if (slides) {
        for (i = 0; i < slides.length; i++) {
            slides[i].style.display = "none";
        }
        slideIndex++;
        if (slideIndex > slides.length) { slideIndex = 1 }
        if (slides[slideIndex - 1]) {
            slides[slideIndex - 1].style.display = "flex";
        }
        setTimeout(() => showSlides(c, slideIndex), 6000); // Change image every 2 seconds
    }
}
