package vyzorix

// ChannelSubscribeRequest.
#ChannelSubscribeRequest: {
        scope: string
}

// ChannelStatusResult.
#ChannelStatusResult: {
        org: string
        active_streams: int
}

// ChannelSubscribeResult.
#ChannelSubscribeResult: {
        subscribed: string
}

// ChannelUnsubscribeResult.
#ChannelUnsubscribeResult: {
        unsubscribed: string
}
