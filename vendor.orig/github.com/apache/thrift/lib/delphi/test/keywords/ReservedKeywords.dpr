program ReservedKeywords;

{$APPTYPE CONSOLE}

uses
  SysUtils, System_;

begin
  try
    { TODO -oUser -cConsole Main : Code hier einf�gen }
  except
    on E: Exception do
      Writeln(E.ClassName, ': ', E.Message);
  end;
end.
