---
title: "Introducción"
date: 2019-03-14T08:44:12+00:00
---

<img src="/banner.png" width="600" alt="Athens Logo"/>

## Athens es un servidor para tus dependencias Go

¡Bienvenidos, Gophers! Estamos deseando explicarte en que consiste Athens...

En esta página web, está documentado Athens en detalle. Te enseñaremos que hace, por que es necesario, y como puedes ejecutarlo por tu cuenta. A continuación te mostramos un breve listado.

#### ¿Qué hace Athens?

Athens proporciona un servidor para [Go Modules](https://github.com/golang/go/wiki/Modules) que puedes ejecutar tú mismo. Almacena código fuente público y privado por tí, para que no tengas que obtenerlo directamente desde un sistema de control de código fuente (VCS) como GitHub o GitLab.

#### ¿Por qué es tan importante? 

Athens actúa como proxy, hay muchas razones por las que querrías tener uno, como por seguridad y velocidad, por poner dos ejemplos. [Echa un vistazo (en inglés)](/intro/why) a algunas de ellas.

#### ¿Cómo lo uso?

Athens es fácil de ejecutar. Te ofrecemos algunas opciones:

- Puedes ejecutarlo como binario en tu sistema
    - Las instrucciones para esto estarán disponibles próximamente 
- Puedes ejecutarlo como imagen de [Docker](https://www.docker.com/) (echa un vistazo [aquí (en inglés)](./install/shared-team-instance/) para saber como hacer esto)
- Puedes ejecutarlo en [Kubernetes](https://kubernetes.io) (echa un vistazo [aquí (en inglés)](./install/install-on-kubernetes/) para saber como hacer esto)

También tenemos una versión experimental de Athens que puedes utilizar sin necesidad de instalar nada. Para utilizarla, establece la variable de entorno `GOPROXY="https://athens.azurefd.net"`.

**[¿Te gusta lo que ves? ¡Prueba athens ahora! (en inglés)](/try-out)**

## ¿Todavía no te ves utilizando Athens?

Aquí te mostramos algunas formas de involucrarte en el proyecto:

* El [listado completo de pasos (en inglés)](/walkthrough) para configurar, ejecutar y testear el proxy Athens explora el procedimiento en mayor detalle.
* ¡Participa en nuestra [reunión semanal de desarrollo (en ingles)](/contributing/community/developer-meetings/)! Es un modo ideal de conocer nuevas personas involucradas en el proyecto, preguntar cosas, o simplemente pasar un buen rato. Todo el mundo es bienvenido a unirse y participar.
* Revisa nuestro listado de [tareas de iniciación](https://github.com/gomods/athens/issues?q=is%3Aopen+is%3Aissue+label%3A%22good+first+issue%22)
* Únete a nosotros en el canal `#athens` en [Gophers Slack](https://invite.slack.golangbridge.org/)

---
El banner de athens ha sido realizado por Golda Manuel
