Contributing to Paho
====================

Thanks for your interest in this project.

Project description:
--------------------

The Paho project has been created to provide scalable open-source implementations of open and standard messaging protocols aimed at new, existing, and emerging applications for Machine-to-Machine (M2M) and Internet of Things (IoT).
Paho reflects the inherent physical and cost constraints of device connectivity. Its objectives include effective levels of decoupling between devices and applications, designed to keep markets open and encourage the rapid growth of scalable Web and Enterprise middleware and applications. Paho is being kicked off with MQTT publish/subscribe client implementations for use on embedded platforms, along with corresponding server support as determined by the community.

- https://projects.eclipse.org/projects/technology.paho

Developer resources:
--------------------

Information regarding source code management, builds, coding standards, and more.

- https://projects.eclipse.org/projects/technology.paho/developer

Contributor License Agreement:
------------------------------

Before your contribution can be accepted by the project, you need to create and electronically sign the Eclipse Foundation Contributor License Agreement (CLA).

- http://www.eclipse.org/legal/CLA.php

Contributing Code:
------------------

The Go client uses git with Gerrit for code review, use the following URLs for Gerrit access;

ssh://<username>@git.eclipse.org:29418/paho/org.eclipse.paho.mqtt.golang

Configure a remote called review to push your changes to;

git config remote.review.url ssh://<username>@git.eclipse.org:29418/paho/org.eclipse.paho.mqtt.golang
git config remote.review.push HEAD:refs/for/<branch>

When you have made and committed a change you can push it to Gerrit for review with;

git push review

See https://wiki.eclipse.org/Gerrit for more details on how Gerrit is used in Eclipse, https://wiki.eclipse.org/Gerrit#Gerrit_Code_Review_Cheatsheet has some particularly useful information.

Git commit messages should follow the style described here;

http://tbaggery.com/2008/04/19/a-note-about-git-commit-messages.html

Contact:
--------

Contact the project developers via the project's "dev" list.

- https://dev.eclipse.org/mailman/listinfo/paho-dev

Search for bugs:
----------------

This project uses Bugzilla to track ongoing development and issues.

- https://bugs.eclipse.org/bugs/buglist.cgi?product=Paho&component=MQTT-Go

Create a new bug:
-----------------

Be sure to search for existing bugs before you create another one. Remember that contributions are always welcome!

- https://bugs.eclipse.org/bugs/enter_bug.cgi?product=Paho
