// Package fakeazerothcore implements a development double of the AzerothCore
// SOAP interface.
//
// It speaks the same wire protocol as the real worldserver (SOAP envelope with
// an executeCommand/commandString request and an executeCommandResponse/result
// reply), keeps the provided account state in memory and logs every command.
//
// Response texts are taken from AzerothCore's `acore_string` table and command
// scripts (cs_account.cpp, cs_ban.cpp, cs_server.cpp) so the double is close to
// the real thing. It is a test aid, not a real server: it never touches MySQL
// and its state is lost on restart unless seeded.
package fakeazerothcore

// Account command response texts, derived from AzerothCore acore_string entries
// and the command scripts.
const (
	msgAccountCreated       = "Account created: %s"
	msgAccountAlreadyExists = "Account with this name already exist!"
	msgAccountNameTooLong   = "Account name can't be longer than 17 characters (client limit), account not created!"
	msgAccountPassTooLong   = "An account password can NOT be longer than 16 characters (client limit). Account NOT created."
	msgAccountNotCreated    = "Account %s NOT created (unknown error)"
	msgAccountNotExist      = "Account not exist: %s"
	msgPasswordChanged      = "The password was changed"
	msgPasswordsDoNotMatch  = "The new passwords do not match"
	msgEmailChanged         = "The email was changed"
	msgEmailsDoNotMatch     = "The new emails do not match"
	msgEmailTooLong         = "Your email can't be longer than 255 characters, email not changed!"
	msgSecurityChanged      = "You change security level of account %s to %d."
	msgLowSecurity          = "You have low security level for this."
	msgInvalidRealmID       = "You have not chosen -1 or the current realmID that you are on."
	msgBadValue             = "Incorrect values."
	msgBanTemporary         = "%s is banned for %s. Reason: %s."
	msgBanPermanent         = "%s is banned permanently. Reason: %s."
	msgBanNotFound          = "%s %s not found"
	msgUnbanned             = "%s unbanned."
	msgUnbanError           = "There was an error removing the ban on %s."
	msgServerUptime         = "Server uptime: %s"
	msgShutdownTimeLeft     = "Time left until shutdown/restart: %s"
	msgCommandNotFound      = "Command '%s' does not exist"
	msgIncorrectSyntax      = "Incorrect syntax."
	msgUnauthorized         = "Authorization failed."
)

const maxCommandBytes = 1 << 20
