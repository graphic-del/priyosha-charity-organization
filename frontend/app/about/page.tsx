export default function AboutPage() {
  return (
    <main className="page-anim">
      <section className="band-maroon page-head">
        <div className="wrap">
          <p className="eyebrow">About</p>
          <h1>About Priyosha</h1>
          <p className="lede">
            A US-based nonprofit raising charity from around the world to reduce
            problems in Bangladesh — for humans and for animals.
          </p>
        </div>
      </section>

      <section className="band">
        <div className="wrap reveal">
          <h2>Our mission</h2>
          <p className="lede">
            To bring support from around the world to Bangladesh, and turn it into
            food, shelter and care — for humans and for animals, especially the cats
            and dogs who suffer the most.
          </p>
          <div className="mission-grid">
            <div><h3>Food</h3><p>Meals and food packs for families who do not have enough to eat.</p></div>
            <div><h3>Shelter</h3><p>Support for people sleeping rough, and a path to more secure housing.</p></div>
            <div><h3>Care for animals</h3><p>Food, shelter and basic veterinary care for stray cats and dogs.</p></div>
          </div>
        </div>
      </section>

      <section className="band-white">
        <div className="wrap reveal">
          <h2>Our founder</h2>
          <div className="founder-card">
            <div className="initials" aria-hidden="true">SM</div>
            <div>
              <h3 style={{ margin: "0 0 6px" }}>Md Shadman Mahmood</h3>
              <p>
                Md Shadman Mahmood founded Priyosha Charity Organization to help the
                people and animals of Bangladesh who suffer the most — people facing
                hunger and sleeping in public places, and the cats and dogs with no
                one to care for them. The organization is based in the United States
                and raises charity from around the world for that work.
              </p>
            </div>
          </div>
          <div className="involve-more">
            <a className="btn" href="/give">Give →</a>
            <a className="btn ghost" href="/contact">Contact us</a>
          </div>
        </div>
      </section>
    </main>
  );
}
