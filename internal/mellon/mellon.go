package mellon

import "fmt"

// Elvish is the Elvish word for "friend" that opens the Doors of Durin.
const Elvish = "mellon"

// Translation is the Common Tongue translation of Elvish.
// Kept as "friend" (English) and "amigo" is handled case-insensitively via IsMellon.
const Translation = "friend"

// PortalInscription is the inscription on the Doors of Durin (in Sindarin/Elvish).
const PortalInscription = "Ennyn Durin Aran Moria: pedo mellon a minno. Im Narvi hain echant: Celebrimbor o Eregion teithant i thiw hin."

// PortalTranslation is the English translation of the inscription.
const PortalTranslation = "The Doors of Durin, Lord of Moria. Speak, friend, and enter. I, Narvi, made them. Celebrimbor of Hollin drew these signs."

// SpeakFriend is the classic riddle answer.
const SpeakFriend = "Speak, friend, and enter."

// Reveal prints the Doors of Durin easter egg.
func Reveal() {
	fmt.Println("✨ _Mellon_ ✨")
	fmt.Println()
	fmt.Println(PortalInscription)
	fmt.Println(PortalTranslation)
	fmt.Println()
	fmt.Println(SpeakFriend)
	fmt.Println()
	fmt.Println("The Doors of Durin open before you. Welcome, friend of Gandalf.")
}

// IsMellon reports whether s is the secret word.
// It accepts "mellon" (Elvish) and "friend" (English). Comparison should be
// done case-insensitively by the caller; this function checks exact lower-case values.
func IsMellon(s string) bool {
	return s == Elvish || s == Translation
}
