ussdproxy
=========

> NOTE: This is still a work in progress and still very experimental!
> I don't even know if it will be worth pursuing in the next 6 months.

`ussdproxy` enables interaction between devices and internet services via USSD; it is especially useful in constrained environments or where internet data costs are high; for example in IOT projects where it may be costly to
acquire internet data for the device to send data to a server. Such projects can leverage modified-UDCP protocol (via USSD) to send data to a central point. This would only require a SIM and the ability to execute a few AT Commands on a GSM modem (...and of course a ussdproxy running and connected to some USSD shortcode)

