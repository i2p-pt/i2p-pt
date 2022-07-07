I2P Pluggable Transport - Streaming Version
===========================================

This is an I2P-based Pluggable Transport for Tor which conceals the
location of an unlisted Tor relay(A "Bridge") and it's client(presumably TBB)
from eachother and from most network observers.

This is a new piece of software and it might or might not be useful. I think it will be,
but I've been wrong before. It was created independently of the I2P project
by an I2P developer. It went through several revisions before this one, and many
of it's features and requirements went into other projects including I2P, it's
libraries, and new projects created to fulfill this purpose. An exhaustive list
of those projects will be added at a later date.

- ***Quick Start** for testers and new users*

- On the *Client-side*:

```md
# Usage (in torrc):
UseBridges 1
Bridge i2p i2pbase32addressesarefiftytwocharacterslongenoughok.b32.i2p
ClientTransportPlugin i2p exec i2p-client
```

- On the *Relay-side*:

```md
# Usage (in torrc):
BridgeRelay 1
ORPort 9001
ExtORPort 7689
ServerTransportPlugin i2p exec i2p-server
```

Why is this useful?
-------------------

I2P and Tor are both different anonymizing technologies, and some might think
that one is objectively superior to the other in the end. I do not believe this
to be the case, I think that I2P and Tor have different advantages and
disadvantages and that the differences between the projects in practice have
to do with the differences between the goals of the developers. If that's the
case, then it pays to think of the networks in terms of their advantages and
disadvantages and consider where they might complement eachother. In particular,
I2P has some advantages that it could offer Tor in situations where those
advantages are useful.

### Among I2P's advantages salient to this point

1. I2P routers can be bootstrapped from any known-good router in the network by
sharing the "RouterInfo" of the known-good router. This allows us to resist
being blocked by sharing the information required to connect by side-channels or
even offline(sneakernet) if need be. Friends who trust eachother can connect to I2P
and get to know I2P routers in other countries without depending on the presence
of bootstrap servers or network consensus.
2. I2P has extremely high relay diversity compared to Tor. Since every participant
in who can safely do so is a relay by default, we have between 60-90 thousand relays
in the network at any given estimate. Each I2P router knows a subset of these relays,
usually between 1500 and 2000, which they use to build their tunnels through. I2P
routers are constantly evaluating the quality of these tunnels in order to make sure
that they are fast and to watch out for potentially colluding routers that might be
conducting attacks. This means that well-integrated I2P routers will only maintain
connections to reliable, fast peers who are not blocked by the regional authority.
3. I2P is capable of onion-routing(Our variant is called garlic routing because we
sometimes bundle messages and pool tunnels, which has a metaphorical clove-like
structure) without Tor's help, and is a hidden service network in it's own right.As
such, it can be used to transport Tor traffic from one Tor client, over I2P, to a
Tor relay, through the Tor network, and out a Tor exit just like any other type of
network connection, but without revealing the IP address of the client to the bridge
and without revealing the IP address of the bridge to the client. It also conceals
the connection from non-colluding network observers who can't conduct a sybil attack.
In other words, it makes it very challenging to enumerate Tor bridges by requesting
them from many addresses/identities to discover and block their location.

### On the other hand, Tor's advantages are

1. Tor has far greater, and far better vetted exit diversity than I2P does. At the time
of this writing, I2P is only recently gained a fleet of exit nodes(which we call
outproxies) which are willing to proxy traffic to the clear web. More importantly,
we have never made any outproxy or pool of outproxies official, we've never offered
legal support to an outproxy operator. Tor's much better at that stuff than us so far
and probably will be for a while, frankly it's not our raison d`etre.
2. Tor allows the use of pluggable transports to connect to a relay in it's network
without revealing the content or destination of communication over that connection to
the operator of that relay. This protects the user of the service from an operator
logging the most specific information about their usage and protects the operator from
being coerced by violence or intimidation into logging specific information, it's
simply materially impossible for them to do so.
3. Tor has a browser bundle. I2P is, at best, a beneficiary of the Tor Browser Bundle's
work in terms of browser privacy and security. How to best mix I2P and Tor in the same
browser is, to my mind, not a straightforward question in spite of the solutions on the
market today. For example:

```md
#### An incomplete overview of problems with mixed I2P/Tor Browsers

- I2P Outproxies that forward `.onion` traffic to Tor
 1. The outproxy can always see what `.onion` sites you are visiting.
 2. It can almost always see the content on you're requestsing.
 3. It does not do anything to address `.onion` sites requesting `.i2p` resources and vice-versa.
- Mixing I2P and Tor with a local proxy(the "Privoxy" method):
 1. Difficult or impossible to prevent `.onion` sites from requesting `.i2p` resources and vice-versa(Mixed-network determination, anonymity set reduction)
 2. Very challenging user-facing implementation
- Mixing I2P and Tor with a proxy.pac file:
 1. Poor browser support and getting poorer
 2. Difficult or impossible to prevent `.onion` sites from requesting `.i2p` resources and vice-versa
- Mixing I2P and Tor with Browser Extensions and Container Tabs:
 1. Relies on non-universal browser features(Container tabs)
 2. Relies on a bunch of software wherein I am the only author(poor peer review)

```

A new way to connect to Tor
---------------------------

All Pluggable Transports are ways of connecting to Tor as long as someone is willing
to provide you a bridge, but this one is the first to use an existing P2P, onion-routed,
obfuscated network to carry traffic to a Tor bridge using the Pluggable Transport
technology. This should use enable Tor to:

 - Avoid bridge enumeration
 - Resist blocking
 - Obfuscate first-hop traffic

A better way forward for mixed I2P/Tor Browsing
-----------------------------------------------

While outproxies will probably always have a place in some threat models, the use of
Tor-over-I2P offers user-facing advantages which I think are worthwhile for I2P users.
This PT Client can be combined with Tor and zzz's [SOCKS outproxy plugin](http://zzz.i2p/topics/3219)
and it will:

 - Route Tor traffic through I2P on it's way to a bridge, preventing Tor traffic from
 leaving the local device unaltered and resisting enumeration. This solves the problem
 of outproxies being able to observe `.onion` traffic.

Combined with a Tor Browser which has been configured for I2P using browser extensions
to intelligently separate applications it can also:

 - Prevent traffic and information in use in `.onion`, clearnet-over-Tor-over-I2P, and
 `.i2p` from leaking information across network and browser-enforced boundaries.
