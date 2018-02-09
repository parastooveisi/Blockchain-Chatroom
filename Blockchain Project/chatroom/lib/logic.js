/**
 * Track the chat of a commodity from one trader to another
 * @param {org.acme.mynetwork.ChatClientChatroom} chat - the trade to be processed
 * @transaction
 */
function chatClientChatroom(chat) {
    return getAssetRegistry('org.acme.mynetwork.ChatMessage')
        .then(function (assetRegistry) {
            return assetRegistry.update(chat.chatMessage);
        });
}

