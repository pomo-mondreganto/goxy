const isDebug = process.env.NODE_ENV === 'development';

let backUrl = '';

if (isDebug) {
    backUrl = 'http://127.0.0.1:8000/api';
} else {
    backUrl = window.location.origin + "/api";
}

export {
    backUrl
}