Fix Before Launch (broken or loses user data)
Campaign Editor

Toolbar buttons steal focus, destroying text selection -- Bold/Italic/Underline silently fail on selected text. Fix: @mousedown.prevent on every toolbar button.
window.prompt for insert link -- loses selection, looks broken
No save confirmation -- spinner disappears, user doesn't know if save worked
Template pill overwrites draft with no confirmation -- data loss
No placeholder text in empty editor

Dashboard / Analytics

Open rate and click rate hardcoded to -- on both views -- actual data never shown
Period selector pills do nothing but look interactive -- remove them or disable + label clearly

Settings -- API Key

Copy button copies bullet characters and shows success checkmark -- actively misleading
Warning text shows before any key exists -- confusing

Setup Wizard

"Continue" disabled with no explanation of why (SMTP test required)
Test email falls back to installation URL string if owner email not set -- sends to a URL

Login

"Run setup wizard" link always visible on a configured system

Fix Before Launch if Quick
Campaign Editor

No active/toggle state on Bold/Italic/Underline/headings
Send Now doesn't validate name/subject/from_email client-side
Schedule modal doesn't validate that a recipient list is selected

Campaigns List

"Date" column label is ambiguous -- different date depending on status
Empty state has no CTA button to create a campaign
Delete scheduled campaign confirm doesn't mention the send will be cancelled

Subscribers

Search ignores active status tab silently
Table shows no loading state while fetching

Lists

No way to edit a list name after creation
Native checkbox for double opt-in toggle in create modal

Subscriber Detail

Tag dropdown has no click-outside to close
"Last Active" always shows --

Modals

Escape key doesn't close modals
Suppression settings uses custom modal instead of BaseModal

Sidebar

No user email or account info shown anywhere

Post-Launch / V2

Full responsive design
Toast / notification system
Replace execCommand with a real rich text library (Tiptap or similar)
Bulk actions on subscribers table
Keyboard shortcuts in editor (Ctrl+B, Ctrl+I)
Page number display in pagination
Tag color picker
Timezone selector with full list
Empty state illustrations
aria-label on tables
Autosave in campaign editor
