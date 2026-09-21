import './App.css';

// The scaffold's demo box is gone: S2-07 replaced the bound surface with the
// real one — Board, Tree, the writes, the habit strip, the timer and the
// settings — and the method it called no longer exists in Go, so its caller
// could not stay either.
//
// This file is deliberately empty of everything else. S2-09 lays out
// frontend/src, S2-11 brings i18n, and the board itself arrives in S2-14 and
// S2-15. Nothing here renders a user-visible string, so there is nothing to
// translate yet.
function App() {
  return <div id="App" />;
}

export default App;
