## Events

This package aims to replace all the internal event handling that is done among different components.
Right now each component declares its callbacks/handlers for each event.

The idea is to have each component sending the events to a common queue and to a common database in
order to add flexibility to the code using an upper layer for handling generic events.

There are two main components, the dispatcher and the workers.

### Dispatcher

The dispatcher main functionality is:

- Check the incoming raw event
- Wrap the event into a shared and generic `Event` struct
- Add the event to the persistent storage for processing

### Workers

There is one worker per event source. One worker is assigned to one source.

_i.e: The events received from the Scrutinizer will be handled one worker respecting a FIFO policy._

#### Available sources

- Scrutinizer
- Keykeeper
- Vochain
- Ethereum
- Census
