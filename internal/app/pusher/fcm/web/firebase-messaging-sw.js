importScripts('https://www.gstatic.com/firebasejs/5.5.9/firebase-app.js');
importScripts('https://www.gstatic.com/firebasejs/5.5.9/firebase-messaging.js');

const SENDER_ID = "713812136942"
// [START initializing]
firebase.initializeApp({
    messagingSenderId: SENDER_ID,
});
// [END initializing]

// [START get_messaging_object]
// Retrieve Firebase Messaging object.
const messaging = firebase.messaging();