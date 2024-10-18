document.addEventListener('DOMContentLoaded', function () {
    var isLoggedIn = document.querySelector('div[data-logged-in]').getAttribute('data-logged-in') === 'true';
    if (!isLoggedIn) {
    setTimeout(function() {
        window.location.href = "/login";
    }, 2000);
}
});