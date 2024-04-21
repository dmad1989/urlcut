/*
Модуль serverapi является серверным слоем сервиса.
Использует github.com/go-chi/chi/v5 в качестве основной http - библитотеки.

Вызовы методов логируются с использованием go.uber.org/zap, поддерживают авторизацию и сжатие данных.
*/
package serverapi