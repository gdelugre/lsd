What is LSD ?
-----------------

**LSD** (for *Lisp Structured Data*) is a data serialization format based on
Lisp s-expressions. It is similar to YAML or JSON but offers a simpler and more
minimalist syntax.

LSD is strongly inspired by self-ml, which was first designed and described on
the `blog of the Chocolat text editor <http://chocolatapp.com/blog/self-ml>`_.
The following implementation is mainly based on this reference with some
extensions of my own, like a few syntactic changes and some conventions to make
it fit into a type system.


Installing
----------

Setup your ``GOPATH`` environment variable (e.g. ``export GOPATH=$HOME/.go``), then run:

.. code-block:: bash

    $ go get github.com/gdelugre/lsd


Usage
-----

.. code-block:: go

    import "github.com/gdelugre/lsd"

    type ConfigExample struct {
        Foo     string 
        Bar     bool
        ... 
    }

    func main() {

        // Initialize structure with default values.
        conf = ConfigExample{...}

        // Use lsd.Load to read a file or lsd.LoadString
        // to parse directly from an existing string.
        if err := lsd.Load("example.lsd", &conf); !err {

            // Structure successfully loaded from file.
            // ...
        } 
    }
    

Syntax
------

**LSD** syntax is based on `s-expressions
<https://en.wikipedia.org/wiki/S-expression>`_ from Lisp, keeping it very light
and simple. This makes it particularly adapted to represent simple
human-readable structures like lists of values or key-value stores.

Let's look for example at this simplified configuration file of an OpenSSH daemon:

.. code-block:: bash

    Port 22
    AddressFamily any
    ListenAddress 0.0.0.0
    HostKey /etc/ssh/ssh_host_dsa_key
    PermitRootLogin no


Here is its counterpart in LSD:

.. code-block:: scheme

    (Port 22)
    (AddressFamily any)
    (ListenAddress 0.0.0.0)
    (HostKey /etc/ssh/ssh_host_dsa_key)
    (PermitRootLogin no)

This example is only composed of key and values. Once parsed it can automatically get mapped to the fields of the following Go structure:

.. code-block:: go

    type SSHConfig struct {
        Port            uint16
        AddressFamily   string
        ListenAddress   string
        HostKey         string
        PermitRootLogin bool
    }

Description
^^^^^^^^^^^

Like self-ml, LSD only recognizes two types of constructs: lists and strings. A
string is a list of UTF-8 characters and a list is a sequence of other lists or
strings. Lists are enclosed by characters ``(`` and ``)``, and their values are
separated by white spaces.

One peculiarity of self-ml lists is that their head *must* be a string. As a
result, empty lists are not permitted.

.. code-block:: scheme

    (MyList this is list (containing (nested lists)) !)

From the language point of view, LSD does not have notion for integers,
booleans or maps as only strings and lists exist. Type checking is nonetheless
present when values are converted into the programming language types (in that
case, Go).

As a result, LSD only focuses on the *structure* of the data and leaves out
the types definition to the reader, allowing it to have a very neat,
uncluttered syntax.

Comments
^^^^^^^^

This implementation differs from `the original self-ml
reference <http://chocolatapp.com/blog/self-ml>`_ as how comments are
represented.  Comments start with a character ``#`` and expands to the end of
the line. To be compatible with most Lisp dialects, the comment character ``;``
can also be used in place of ``#``. This form is preferred as it naturally
allows better syntax highlighting.

Multi-line block comments are not supported.

.. code-block:: scheme

    ;
    ; This is a comment.
    ; 
    (Persons John Jane William Sarah)

Booleans and numbers
^^^^^^^^^^^^^^^^^^^^

Boolean values can be represented by the following strings: ``0``, ``1``, ``true``, ``false``, ``yes`` and ``no``. Capitalized and upper-case
versions are also accepted.

Integers are written in decimal by default. A few common notations are also recognized:

  * *hexadecimal* when the string is prefixed by ``0x``
  * *binary* when the string is prefixed by ``0b``
  * *octal* when the string starts with a ``0``

Escaped strings
^^^^^^^^^^^^^^^

Strings may need to contain white spaces or special characters like ``(``,
``)`` or ``;``. Two possible methods exist to define such strings.

* Quoted strings

  Strings are enclosed between ``"`` (double quote) characters, and everything
  between those delimitors is part of the string. If the string also contains
  ``"`` characters, those must be doubled. Backticks (`````) can also be used
  in place of double quotes.

  .. code-block:: scheme

    (Description "This is a long ""string"" with whitespaces")

* Bracketed strings

  Strings are enclosed between ``[`` and ``]`` and everything between brackets
  is part of the string. The string can also contain brackets as long as they
  are balanced.

  .. code-block:: scheme

    (Description [This is a bracketed string [with other brackets]])

Structures
^^^^^^^^^^

Structures can be seen and constructed in two different ways:

  * As a key-value dictionary for which keys are fixed and correspond to field names
  * As a sequence of values in a specific order

When a structure is constructed by field names, every values must be sub-lists
for which heads correspond to the field names of the structure.

For instance, let's consider the following declaration in Go:

.. code-block:: go

    type PlayerInfo struct {
        UserName     string
        CurrentLevel int
        Score        float32
    }

This structure could be initialized by field name in the following way:

.. code-block:: scheme

    (UserName acidburn)
    (CurrentLevel 2)
    (Score 133.7)

This structure could as well be nested in another structure:

.. code-block:: go

    type PlayerProfile struct {
        Info    PlayerInfo
    }

.. code-block:: scheme

    ; Struct definition by field name.
    (Info
        (UserName acidburn)
        (CurrentLevel 2)
        (Score 133.7))
    
If you want to define it by field order, then simply put the values in the order they appear in the structure:

.. code-block:: scheme

    ; Struct definition by field order.
    (Info acidburn 2 59.14)

Any structure can be constructed by name or by order. The only exception is the
root structure that must be created using field names (which correspond to
top-level list definitions in a LSD document).

Maps
^^^^

Maps are key-value dictionaries in Go. They can be defined in the same fashion
we define a structure by its fields name. Unlike structs, the keys can be of
any value as long as they can be converted into their Go native type.

Since list heads must be defined as string values in LSD, the key type of
the map *must not* be of a compound-type (like struct, map or slice).

.. code-block:: scheme

    ; field Options has type map[string]bool
    (Options
        (EnableFeatureX true)
        (EnableFeatureY false)
        (EnableFeatureZ true)
    )

Slices and arrays
^^^^^^^^^^^^^^^^^

Slices are variable-length arrays and are naturally represented by lists:

.. code-block:: scheme

    ; KnownHackers []string
    (KnownHackers acidburn zerocool crashoverride)

Arrays follow the same convention with the additional constraint that the
number of values must not overflow the length of the array.

Since LSD only allows to define strings for list heads, one problem may
arise you try to create a list of a compound type. If you define a slice of
slices, you can define an empty string for the sub-list head. The recommended
notation is ``[]``, which is self-talkative for a list:

.. code-block:: scheme

    ; Measures [3][3]float32
    (Measures
        ([] 1.02 4.29 0.12)
        ([] 0.00 1.20 4.40)
        ([] 3.43 1.11 4.85)
    )

Another possibility is to use a **bullet point** to mark the beginning of the
list.  The allowed bullets are: ``-``,  ``*``, ``•`` (U+2022), ``‣`` (U+2023),
``⁃`` (U+2043) and ``◦`` (U+25E6).

Consider the following declaration in Go:

.. code-block:: go

    type User struct {
        UserName string
        Age      uint
        Email    string
        Admin    bool
    }

    type RegisteredUsers struct {
        Users []User
    }

Here is the definition is LSD:

.. code-block:: scheme

    (Users
        (‣ (UserName root) (Admin true))
        (‣ (UserName Emma) (Age 27) (Email emma@example.com) (Admin false)) 
        (‣ (UserName Josh) (Age 32) (Email josh@example.com) (Admin false)) 
    )

Example of a LSD file
---------------------

.. code-block:: scheme

    ;
    ; Simple init script in LSD.
    ;

    ; Generic description.
    (Name exampledaemon)
    (Description "Does plenty of stuff")

    ; Runtime information.
    (Type standalone)
    (PidFile /var/run/daemon.pid)
    (Security
        (Chroot /var/run/daemon/ 0755)
        (User nobody) (Group nobody)
        (Capabilities
            Net
            Log)
    )

    ; Start/Stop configuration.
    (Environment /etc/conf.d/daemon)
    (Handlers
        (start "/usr/bin/daemon --quiet")
        (stop "/usr/bin/kill -TERM $DAEMONPID")
        (reload "/usr/bin/kill -HUP $DAEMONPID")
    )
    (StopTimeout 10.000)


    ; Service dependencies.
    (Depends
        network
        syslog
        localfs)

Notes
-----

This specification has primarily been thought for my own needs as I was
looking for an elegant and minimalist configuration file format. This can be
used to deserialize any kind of data though. I'd be very pleased if you have
remarks or suggestions.

Feel free to reach me at ``<guillaume AT security-labs DOT org>``.

