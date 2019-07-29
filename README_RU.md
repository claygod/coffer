# Coffer

Простая key-value ACID* база данных.

*is a set of properties of database transactions intended to guarantee validity even in the event of errors, power failures, etc.

### Старт

ПРи старте последним по номеру должен быть чекпойнт. Если это не так, то значит, остановка была некорректной.
Тогда грузится последний имеющийся чекпоинт и все логи после него до тех пор, пока это возможно. На битом логе
или последнем логе скачиваем, пока получается, и на этом загрузку заканчиваем. БД создаёт новый чекпоинт,
и после этого возможно продолжение исполнение кода.

### Copyright © 2019 Eduard Sesigin. All rights reserved. Contacts: <claygod@yandex.ru>
