package bot

const API string = "https://discord.com/api/v9"
const WS_API string = "wss://gateway.discord.gg"
const CHANNEL_TEXT uint8 = 0
const CHANNEL_VOICE uint8 = 2
const CHANNEL_TEXT_CATEGORY uint8 = 4

const CHAT_INPUT uint8 = 1

const WS_DISPATCH uint8 = 0
const WS_HEARTBEAT uint8 = 1
const WS_INDENTIFY uint8 = 2
const WS_STATUS_UPDATE uint8 = 3
const WS_VOICE_STATE_UPDATE uint8 = 4

const ARG_SUBCOMMAND uint8 = 1
const ARG_SUBCOMMAND_GROUP uint8 = 2
const ARG_STRING uint8 = 3
const ARG_INTEGER uint8 = 4
const ARG_BOOLEAN uint8 = 5
const ARG_USER uint8 = 6
const ARG_CHANNEL uint8 = 7
