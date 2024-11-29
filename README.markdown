# Ahm

> ⚠️ This is experimental software.
> It it prone to change.
>
> Feel free to play around.
> Think twice before relying on it for anything that needs to be rock-solid.

Ahm is a minimalist, lightweight markup format.
The name stands for _at-sign header markup_.

## Goals

* **Short pick-up time.**
  A busy passer-by is able to read the syntax with little guesswork.
  They have a high chance of writing it correctly after seeing a sample.

* **Simplicity and consistency.**
  An unambigous syntax that does not admit weird and complex edge cases.
  Easy to implement, with different implementations working the same.

* **Low ceremony.**
  Few demands are placed on the user.

* **Generality and extensibility.**
  A syntax without semantics that can be used for different purposes.
  Easy to process, validate, and constrain programmatically.

## Comparisons

These are not _all_ the languages one could compare Ahm with.
The list is not supposed to be exhaustive.
The ones listed are chosen because of their prevalence and relation to Ahm's goals.

* **Markdown** is a markup language supported in many tools.
  It is generally easy to get started with, but hides gnarly complexities.

  Ambiguities in the original description lead to many incompatible implementations.
  CommonMark is an admirable effort to produce an unambigous version.
  The specifications remain fairly complex and time consuming to implement.

  It is not extensible.
  Static site generators using Markdown often add support for a Yaml front-matter.
  This is done outside of the markup language itself and requires mixing different tools.

* **HTML** and **XML** are heavyweight markup languages.
  Their basic syntaxes are very close to each other.

  HTML is more immediately usable for document preparation.
  It is not a format most people would choose for structured or semi-structured notes.
  It is not readable nor terse enough for those.

  XML is more extensible and better suited for structured data.
  It has a family of related technologies (XSLT, XPath, XQuery, XSD).
  While it is by no means dead, it's popularity _has_ declined.

* **S-expressions** are the basic notation of the Lisp programming language family.
  They're not markup, but for programming languages, they achieve seveal of the Ahm goals.

## Samples

A to-do list.

```ahm
@DONE Buy milk
@TODO Get recipe
@TODO Bake cake
```

A script for a [storylet]-based narrative game.

[storylet]: https://emshort.blog/2019/11/29/storylets-you-want-them/

```ahm
@EPISODE A New Hope
  
  @CONDITIONS
    @VALUE Luke.yearning-to-leave 40
    @VALUE Force.disturbance 10
  
  @CARD title-screen
    A long time ago, in a Galaxy not so far from here.
    
    *Pompous music is playing.*
    
    @CHOICE
      @OPTION "Change the music theme to Star Trek!"
        @EFFECT
          ...

      @OPTION "Change the music theme to Doctor Who!"
        @EFFECT
          ...
```

A document.

```ahm
@DOCUMENT
  @TITLE A quick and dirty documentation format
  @AUTHOR Karol Marcjan
  @DATE
    @TODAY

@IMPORT GO

This is how you'd do something Markdown-like in Ahm.

@H1 A header.
@H2 A nested header.
@H6 HTML goes up to this, right?

Kinetic energy is
@MATH E_k = \frac{1}{2} m v^2
, as you should remember from school.

Want a code block? You can have a code block!

@CODE go
  @GO/FMT
    package main

    import "fmt"

    func main() {
      fmt.Println("Hello, world!")
    }

Want to include from an outside file?
@CODE go
  @INCLUDE-FROM ../somedir/some_file.go
    @GO/FUNCTION-BODY someExampleFunc
```

## Copyright

Ahm ©️ Karol Marcjan 2024